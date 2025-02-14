resource "scaffolding_publisher" "pub1" {
  description = "my-publisher"
}

resource "scaffolding_book" "book1" {
  price = "1"  
  publisher = scaffolding_publisher.pub1.path
  published = true
  isbn = ["1234"]
}
