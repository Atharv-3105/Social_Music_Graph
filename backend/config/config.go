package config

import (
	"os"
	"github.com/joho/godotenv"
)


type Config struct {
	Neo4jURI   		string 
	Neo4jUser		string 
	Neo4jPassword	string 
}

func Load() *Config {

	//Load .env file
	godotenv.Load()

	return &Config {
		Neo4jURI:        getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser: 		 getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:   getEnv("NEO4J_PASSWORD", "password"),	
	}
}


func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return fallback
}