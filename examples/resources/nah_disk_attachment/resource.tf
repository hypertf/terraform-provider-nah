resource "nah_disk_attachment" "data" {
  project     = nah_project.app.slug
  disk_id     = nah_disk.data.id
  instance_id = nah_instance.app.id
  device      = "vdb"
}
