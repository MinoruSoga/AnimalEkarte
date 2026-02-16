terraform {
  backend "s3" {
    bucket         = "animalekarte-tfstate-698109622668"
    key            = "env/test/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "animalekarte-terraform-lock"
    encrypt        = true
  }
}
