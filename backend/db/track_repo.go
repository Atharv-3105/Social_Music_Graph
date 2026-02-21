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

// func GetMusicSimilarity(userID string) ([]map[string]any, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
// 	defer cancel()

// 	session := Driver.NewSession(ctx, neo4j.SessionConfig{
// 		AccessMode: neo4j.AccessModeRead,
// 	})
// 	defer session.Close(ctx)

// 	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

// 		query := `
// 				MATCH (u1:User {user_id: $user_id})
// 				MATCH (u1)-[:LISTENS]->(t:Track)<-[:LISTENS]-(u2:User)
// 				WHERE u1 <> u2
				
// 				RETURN us.user_id AS similar_user,
// 					   COUNT(t) AS common_tracks
// 				ORDER BY common_tracks DESC
// 				LIMIT 10
// 			`
		
// 		res, err := tx.Run(ctx, query, map[string]any{
// 			"user_id" : userID,
// 		})
// 		if err != nil {
// 			return nil, err
// 		}

// 		results := []map[string]any{}

// 		for res.Next(ctx) {
// 			record := res.Record()


// 			results  = append(results, map[string]any{
// 				"user_id" : record.Values[0],
// 				"common_tracks" : record.Values[1],
// 			})
// 		}

// 		return results, nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	return result.([]map[string]any) , nil
// }

func GetJaccardSimilarity(userID string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	//We need to perform UNION,INTERNSECTION inorder to Compute Jaccard Similarity
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

		query := `
				MATCH (u1: User {user_id : $user_id})
				MATCH (u1)-[:LISTENS]->(t:Track)<-[:LISTENS]-(u2:User)
				WHERE u1 <> u2
				
				
				WITH u1, u2, count(DISTINCT t) AS intersection
				
				MATCH (u1)-[:LISTENS]->(t1:Track)
				WITH u1, u2, intersection, count(DISTINCT t1) AS total1
				
				MATCH (u2)-[:LISTENS]->(t2: Track)
				WITH u2.user_id AS similar_user,
							intersection,
							total1,
							count(DISTINCT t2) AS total2
				
				WITH similar_user,intersection,(total1 + total2 - intersection) AS union_count
				
				WHERE union_count > 0
				
				RETURN similar_user,
					   intersection,
					   union_count,
					   toFloat(intersection) / union_count AS jaccard
				
				ORDER BY jaccard DESC
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

					results = append(results, map[string]any{
						"user_id" : record.Values[0],
						"intersection" : record.Values[1],
						"union" : record.Values[2],
						"jaccard" : record.Values[3],
					})
				}

				return results, nil
	})

	if err != nil {
		return nil, err 
	}

	return result.([]map[string]any) , nil
}


func GetCosineSimilarity(userID string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

		query := `
				MATCH (u1 : User {user_id : $user_id})-[r1:LISTENS]->(t:Track)
				MATCH (u2 : User)-[r2:LISTENS]->(t)
				WHERE u1 <> u2
				
				WITH u1, u2, 
					  SUM(r1.count * r2.count) AS dot_product
					  
				MATCH (u1)-[r:LISTENS]->(:Track)
				WITH u2, dot_product,
				 	 sqrt(SUM(r.count * r.count)) AS norm1
				
				MATCH (u2)-[r:LISTENS]->(:Track)
				WITH u2, dot_product,
					  norm1,sqrt(SUM(r.count * r.count)) AS norm2
				
				WITH similar_user,
					 dot_product,
					 norm1,norm2,
					 CASE 
					 	WHEN norm1 = 0 OR norm2 = 0
						THEN 0
						ELSE dot_product / (norm1 * norm2)
					 END AS cosine
					 
				RETURN similar_user,
					   dot_product,
					   cosine
				ORDER BY cosine DESC
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

			results = append(results, map[string]any{
				"user_id" : record.Values[0],
				"dot_product" : record.Values[1],
				"cosine" : record.Values[2],
			})
		}

		return results, nil
	})	

	if err != nil {
		return nil, err
	}

	return result.([]map[string]any), nil
}