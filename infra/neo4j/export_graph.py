from neo4j import GraphDatabase
import torch 
import numpy as np 
import pandas as pd 
import os
from sklearn.preprocessing import StandardScaler

from dotenv import load_dotenv
load_dotenv()

URI = os.getenv("NEO4J_URI")
USER = os.getenv("NEO4J_USER")
PASSWORD = os.getenv("NEO4J_PASSWORD")

driver = GraphDatabase.driver(URI, auth = (USER, PASSWORD))

def fetch_users(tx):
    result = tx.run("MATCH (u:User) RETURN u.user_id AS id")
    return [record["id"] for record in result]


def fetch_user_similarity_edges(tx):
    query = """
    MATCH (u1 : User)-[:LISTENS]->(t:Track)<-[:LISTENS]-(u2:User)
    WHERE u1.user_id < u2.user_id
    
    WITH u1, u2, count(DISTINCT t) AS intersection
    
    MATCH (u1)-[:LISTENS]->(t1:Track)
    WITH u1, u2, intersection, count(DISTINCT t1) AS total1
    
    MATCH(u2)-[:LISTENS]->(t2:Track)
    WITH u1, u2, intersection, total1, 
         count(DISTINCT t2) AS total2
         
    WITH u1, u2,
        intersection,
        (total1 + total2 - intersection) AS union_count
    
    WITH u1.user_id AS u1,
         u2.user_id AS u2,
         toFloat(intersection) / union_count AS jaccard
        
    WHERE jaccard > 0.05
    
    RETURN u1, u2, jaccard AS weight
    """
    result = tx.run(query)
    return [(r["u1"], r["u2"], r["weight"]) for r in result]


#Fetch features such as Total_Listens, Unique_tracks, Listening Entropy(how diverse is the taste of the user)
def fetch_user_features(tx):
    query = """
        MATCH (u: User)-[r:LISTENS]->(t:Track)
        WITH u,
             SUM(r.count) AS total_listens,
             COUNT(DISTINCT t) AS unique_tracks,
             COLLECT(r.count) AS counts
        
        WITH u,
             total_listens,
             unique_tracks,
             counts,
             [c IN counts | toFloat(c) / total_listens] AS probs
        
        WITH u,
             total_listens,
             unique_tracks,
             REDUCE(entropy = 0.0, p IN probs |
                    entropy - (p * log(p))
            ) AS entropy
        
        RETURN u.user_id AS user_id,
               total_listens,
               unique_tracks,
               entropy
    """
    result = tx.run(query)
    return [record.data() for record in result]

with driver.session() as session:
    print("fetching users...")
    users = session.execute_read(fetch_users)
    
    print("fetching similarity edges...")
    edges = session.execute_read(fetch_user_similarity_edges)
    
    print("fetching user features....")
    features_data = session.execute_read(fetch_user_features)
    
driver.close()


#Map the Users to Indices
user_to_idx = {u : i for i, u in enumerate(users)}
num_users = len(users)

#Build EDGE INDEX
edge_index = []
edge_weight = []

for u1,u2, weight in edges:
    i = user_to_idx[u1]
    j = user_to_idx[u2]
    
    #UNDIRECTED GRAPH
    edge_index.append([i,j])
    edge_index.append([j, i])
    
    edge_weight.append(weight)
    edge_weight.append(weight)

edge_index = torch.tensor(edge_index, dtype = torch.long).t().contiguous()
edge_weight = torch.tensor(edge_weight, dtype = torch.float)

print("Users:", num_users)
print("Edges: ", edge_index.shape)

#Basic Node Features
features = np.zeros((num_users, 3))

#Compute Degreee
degree_count = {}
for u1, u2 , _  in edges:
    degree_count[u1] = degree_count.get(u1, 0) + 1
    degree_count[u2] = degree_count.get(u2, 0) + 1
    
for row in features_data:
    uid = row["user_id"]
    idx = user_to_idx[uid]
    
    features[idx, 0] = row["total_listens"]
    features[idx, 1] = row["unique_tracks"]
    features[idx, 2] = row["entropy"]

#Normalize features before converting them into Tensors
scaler = StandardScaler()
features = scaler.fit_transform(features)
 
x = torch.tensor(features, dtype = torch.float)

torch.save({
    "x" : x,
    "edge_index" : edge_index,
    "edge_weight" : edge_weight,
    "user_to_idx" : user_to_idx
}, "graph_data2.pt")

print("Graph export complete.")