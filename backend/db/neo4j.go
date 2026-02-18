package db

import (
	"context"
	"time"

	"github.com/atharv-3105/Social_Music_Graph/logger"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var Driver neo4j.DriverWithContext


func Init(uri, user, pswrd string) error {
	driver, err := neo4j.NewDriverWithContext(uri, 
					neo4j.BasicAuth(user, pswrd, ""))

	if err != nil {
		logger.Log.WithError(err).Error("Failed to create Neo4j Driver")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		logger.Log.WithError(err).Error("Neo4j connectivity failed")
		return err
	}


	Driver = driver
	logger.Log.Info("Connected to Neo4j successfully")
	return nil
}