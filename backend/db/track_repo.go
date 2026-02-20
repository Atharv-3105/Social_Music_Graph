package db

import (
	"context"
	"time"

	"github.com/atharv-3105/Social_Music_Graph/logger"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)



func AddListen(userID, trackID, title string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

		query := `
				MATCH (u : User {user_id : $user_id})
				MERGE (t : Track {track_id : $track_id})
				ON CREATE SET t.title = $title
				
				MERGE (u)-[r:LISTENS]->(t)
				ON CREATE SET r.count = 1
				ON MATCH SET r.count = r.count + 1
				
				RETURN r.count
			`
		
		res, err := tx.Run(ctx, query, map[string]any{
			"user_id" : userID,
			"track_id" : trackID,
			"title" : title,
		})
		if err != nil {
			return nil, err
		}

		if res.Next(ctx) {
			record := res.Record()
			count := record.Values[0]

			logger.Log.WithFields(map[string]any{
				"user_id" : userID,
				"track_id" : trackID,
				"count" : count,
			}).Info("listen recorded")

			return count, nil
		}

		return nil, nil
	})

	return err
}

func GetMusicSimilarity(userID string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

		query := `
				MATCH (u1:User {user_id: $user_id})
				MATCH (u1)-[:LISTENS]->(t:Track)<-[:LISTENS]-(u2:User)
				WHERE u1 <> u2
				
				RETURN us.user_id AS similar_user,
					   COUNT(t) AS common_tracks
				ORDER BY common_tracks DESC
				LIMIT 10
			`
		
		res, err := tx.Run(ctx, query, map[string]any{
			"user_id" : userID,
		})
		if err != nil {
			return nil, err
		}

		results := []map[string]any{}

		for res.Next(ctx) {
			record := res.Record()


			results  = append(results, map[string]any{
				"user_id" : record.Values[0],
				"common_tracks" : record.Values[1],
			})
		}

		return results, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]map[string]any) , nil
}