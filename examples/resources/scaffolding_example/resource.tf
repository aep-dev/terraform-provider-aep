provider "scaffolding" {
  headers = {
    "content-type" = "application/json",
    "x-api-key" = var.RobloxApiKey
  }
}

variable "RobloxApiKey" {

}

variable "data_store_entries" {
  type = list(object({
    id = string
    value = string
  }))
  default = [
    { id = "id1", value = "value1" },
    { id = "id2", value = "value2" },
    { id = "id3", value = "value3" },
    { id = "id4", value = "value4-2" },
    { id = "id5", value = "value5-2" },
    { id = "id6", value = "value6-2" },
    { id = "id7", value = "value7" },
    { id = "id8", value = "value8" },
    { id = "id9", value = "value9" },
    { id = "id10", value = "value10" }
  ]
}

resource scaffolding_data-store-entry "ds" {
  count = length(var.data_store_entries)

  universe_id = "7308449638"
  data_store_id = "demo-1"
  id = var.data_store_entries[count.index].id
  value = var.data_store_entries[count.index].value
}

resource scaffolding_data-store-entry "ds2" {
  count = length(var.data_store_entries)

  universe_id = "7308449638"
  data_store_id = "demo-2"
  id = var.data_store_entries[count.index].id
  value = var.data_store_entries[count.index].value
}