package db

import (
	"context"
	"errors"
	"time"

	"github.com/atharv-3105/Social_Music_Graph/logger"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserNotFound  = errors.New("user not found")
var ErrAlreadyFollowing = errors.New("already following")
var ErrCannotFollowSelf = errors.New("cannot follow self")

func CreateUser(userID, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()


	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite, //Write mode
	})
	defer session.Close(ctx)

	_ , err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error){
		query :=`
				CREATE (u:User {user_id: $user_id, name: $name})
				RETURN u
				`
		_, err := tx.Run(ctx, query, map[string]any{
			"user_id": userID,
			"name":    name,
		})
		return nil, err
	})
		
	if err != nil {

		//Detect Neo4j constraint error
		if neoErr, ok := err.(*neo4j.Neo4jError); ok {
			if neoErr.Code == "Neo.ClientError.Schema.ConstraintValidationFailed"{
				logger.Log.WithField("user_id", userID).Warn("user already exists")
				return ErrUserAlreadyExists
			}
		}
		logger.Log.WithError(err).Error("failed to create user in DB")
		return err
	}

	logger.Log.WithField("user_id", userID).Info("user created successfully")
	return nil
}


func GetUser(userID string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead, //Read Mode
	})

	defer session.Close(ctx)


	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
				MATCH (u : User {user_id: $user_id})
				RETURN u.user_id AS user_id, u.name as name`
		
		res, err := tx.Run(ctx, query, map[string]any{
			"user_id": userID,
		})
		if err != nil {
			return nil, err
		}

		if res.Next(ctx){
			record := res.Record()

			return map[string]any{
				"user_id": record.Values[0],
				"name":    record.Values[1],
			}, nil
		}

		return nil, ErrUserNotFound
	})

	if err != nil {
		return "", "", err
	}

	data := result.(map[string]any)

	return data["user_id"].(string), data["name"].(string), nil
}


func FollowUser(followerID, followeeID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error){
		query := `  
				MATCH (u1 : User {user_id : $follower_id})
				MATCH (u2 : User {user_id : $followee_id})

				OPTIONAL MATCH (u1)-[r:FOLLOWS]->(u2)

				WITH u1,u2,r 
				CALL{
						WITH u1, u2, r
						RETURN CASE
								WHEN u1 IS NULL OR u2 IS NULL THEN "USER_NOT_FOUND"
								WHEN r IS NOT NULL THEN "ALREADY_FOLLOWING"
								ELSE "CREATE"
						END AS action
					}

				WITH u1, u2, action 
				WHERE action = "CREATE"
				CREATE (u1)-[:FOLLOWS]->(u2)

				RETURN action
		` 

		res, err := tx.Run(ctx, query, map[string]any{
			"follower_id": followerID,
			"followee_id": followeeID,
		})
		if err != nil {
			return nil, err
		}

		if res.Next(ctx){

			record := res.Record()
			action := record.Values[0].(string)

			switch action {
			case "CREATE":
						return nil, nil
			case "ALREADY_FOLLOWING":
						return nil, ErrAlreadyFollowing
			case "USER_NOT_FOUND":
						return nil, ErrUserNotFound
			}
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	logger.Log.WithFields(map[string]any{
		"follower": followerID,
		"followee": followeeID,
	}).Info("follow relationship created")

	return nil
}


func GetFollowers(userID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

		query := `
				MATCH (u:User (user_id: $user_id))
				OPTIONAL MATCH (f:User)-[:FOLLOWS]->(u)
				RETURN u, collect(f.user_id) AS followers
			`
			
		res, err := tx.Run(ctx, query, map[string]any{
			"user_id":userID,
		})
		if err != nil {
			return nil, err
		}

		if res.Next(ctx){
			record := res.Record()

			if record.Values[0] == nil {
				return nil, ErrUserNotFound
			}

			followers := []string{}

			raw := record.Values[1]
			if raw != nil {
				for _, v := range raw.([]any) {
					if v != nil {
						followers = append(followers, v.(string))
					}
				}
			}
			return followers, nil
		}
		return nil, ErrUserNotFound
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

func GetFollowing(userID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

		query := `
				MATCH (u: User {user_id: $user_id})
				OPTIONAL MATCH (u)-[:FOLLOWS]->(f:User)
				RETURN u, collect(f.user_id) AS following
		`
		
		
		res, err := tx.Run(ctx, query, map[string]any{
			"user_id": userID,
		})

		if err != nil {
			return nil, err
		}

		if res.Next(ctx){
			record := res.Record()

			if record.Values[0] == nil {
				return nil, ErrUserNotFound
			}

			following := []string{}

			raw := record.Values[1]
			if raw != nil {
				for _, v := range raw.([]any) {
					if v != nil {
						following = append(following, v.(string))
					}
				}
			}

			return following, nil
		}

		return nil, ErrUserNotFound
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}


func SuggestUsers(userID string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	session := Driver.NewSession(ctx, neo4j.SessionConfig {
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

		query := ` 
				 MATCH (u : User {user_id: $user_id})
				 MATCH (u)-[:FOLLOWS]->(f1)-[:FOLLOWS]->(candidate)
				 WHERE candidate.user_id <> $user_id
				 AND NOT (u)-[:FOLLOWS]->(candidate)
				 
				 RETURN candidate.user_id AS user_id,
				 		COUNT(f1) AS mutual_count
				 ORDER BY mutual_count DESC
				 LIMIT 10
			`

		res, err := tx.Run(ctx, query, map[string]any{
			"user_id": userID,
		})

		if err != nil {
			return nil, err
		}

		suggestions := []map[string]any{}

		for res.Next(ctx) {
			record := res.Record()

			suggestions = append(suggestions, map[string]any{
				"user_id":		record.Values[0],
				"mutual_count": record.Values[1],
			})
		}

		return suggestions, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]map[string]any), nil
}