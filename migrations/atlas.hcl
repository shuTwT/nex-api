variable "database_url" {
	type    = string
	default = getenv("DATABASE_URL")
}

env "local" {
	url = var.database_url
	dev = "sqlite://file?mode=memory&_fk=1"

	migration {
		dir = "file://migrations"
	}
}
