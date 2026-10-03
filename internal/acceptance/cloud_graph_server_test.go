package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// graphContractServer implements the frozen cloud-graph contract in memory. It
// supplements the pinned upstream server without contacting any deployment.
type graphContractServer struct {
	next uint64
	mu   sync.Mutex
	data map[string]map[string]any
	base http.Handler
}

func newGraphContractServer(base http.Handler) http.Handler {
	return &graphContractServer{data: map[string]map[string]any{}, base: base}
}

func (s *graphContractServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	kind, scope, id, collection, project, parent := graphRoute(parts)
	if kind == "" {
		s.base.ServeHTTP(w, r)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	key := kind + "/" + scope + "/" + id
	switch r.Method {
	case http.MethodPost:
		if !collection {
			http.NotFound(w, r)
			return
		}
		var value map[string]any
		if json.NewDecoder(r.Body).Decode(&value) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		s.next++
		id = fmt.Sprintf("%032x", s.next)
		key = kind + "/" + scope + "/" + id
		value["id"] = id
		if project != "" {
			value["project_id"] = "000000000000000000000000000000aa"
		} else {
			value["org_id"] = "000000000000000000000000000000bb"
		}
		switch kind {
		case "subnet":
			value["network_id"] = parent
		case "disk_attachment":
			value["disk_id"] = parent
		case "policy_binding":
			value["policy_id"] = parent
		case "load_balancer_backend":
			value["load_balancer_id"] = parent
		}
		if kind == "load_balancer" {
			value["region"] = "eu-west-1"
			value["status"] = "active"
		}
		if kind == "load_balancer_backend" {
			value["healthy"] = value["enabled"]
		}
		now := time.Now().UTC().Format(time.RFC3339)
		value["created_at"], value["updated_at"] = now, now
		s.data[key] = value
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(value)
	case http.MethodGet:
		value, ok := s.data[key]
		if collection || !ok {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(value)
	case http.MethodPatch:
		value, ok := s.data[key]
		if collection || !ok {
			http.NotFound(w, r)
			return
		}
		if kind == "disk_attachment" || kind == "policy_binding" {
			w.WriteHeader(405)
			return
		}
		var patch map[string]any
		if json.NewDecoder(r.Body).Decode(&patch) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		if kind == "disk" && number(patch["size_gb"]) < number(value["size_gb"]) {
			http.Error(w, "disk shrinking is not supported", 400)
			return
		}
		for k, v := range patch {
			value[k] = v
		}
		if kind == "load_balancer_backend" {
			value["healthy"] = value["enabled"]
		}
		value["updated_at"] = time.Now().UTC().Format(time.RFC3339)
		_ = json.NewEncoder(w).Encode(value)
	case http.MethodDelete:
		if collection {
			http.NotFound(w, r)
			return
		}
		if _, ok := s.data[key]; !ok {
			http.NotFound(w, r)
			return
		}
		delete(s.data, key)
		w.WriteHeader(204)
	default:
		w.WriteHeader(405)
	}
}
func number(v any) float64 { n, _ := v.(float64); return n }

func graphRoute(p []string) (kind, scope, id string, collection bool, project, parent string) {
	if len(p) >= 2 && p[0] == "v1" && p[1] == "policies" {
		if len(p) == 2 {
			return "policy", "org", "", true, "", ""
		}
		if len(p) == 3 {
			return "policy", "org", p[2], false, "", ""
		}
		if len(p) >= 4 && p[3] == "bindings" {
			scope = "org/" + p[2]
			if len(p) == 4 {
				return "policy_binding", scope, "", true, "", p[2]
			}
			if len(p) == 5 {
				return "policy_binding", scope, p[4], false, "", p[2]
			}
		}
		return
	}
	if len(p) < 4 || p[0] != "v1" || p[1] != "projects" {
		return
	}
	project = p[2]
	collectionName := p[3]
	if collectionName == "networks" && len(p) >= 6 && p[5] == "subnets" {
		scope = project + "/" + p[4]
		if len(p) == 6 {
			return "subnet", scope, "", true, project, p[4]
		}
		if len(p) == 7 {
			return "subnet", scope, p[6], false, project, p[4]
		}
	}
	if collectionName == "disks" && len(p) >= 6 && p[5] == "attachments" {
		scope = project + "/" + p[4]
		if len(p) == 6 {
			return "disk_attachment", scope, "", true, project, p[4]
		}
		if len(p) == 7 {
			return "disk_attachment", scope, p[6], false, project, p[4]
		}
	}
	if collectionName == "load-balancers" && len(p) >= 6 && p[5] == "backends" {
		scope = project + "/" + p[4]
		if len(p) == 6 {
			return "load_balancer_backend", scope, "", true, project, p[4]
		}
		if len(p) == 7 {
			return "load_balancer_backend", scope, p[6], false, project, p[4]
		}
	}
	kinds := map[string]string{"networks": "network", "disks": "disk", "load-balancers": "load_balancer", "instances": "instance"}
	kind = kinds[collectionName]
	scope = project
	if kind == "" {
		return
	}
	if len(p) == 4 {
		return kind, scope, "", true, project, ""
	}
	if len(p) == 5 {
		return kind, scope, p[4], false, project, ""
	}
	return "", "", "", false, "", ""
}
