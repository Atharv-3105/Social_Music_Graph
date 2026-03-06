import numpy as np
from neo4j import GraphDatabase
from tqdm import tqdm 
import random
import os
from dotenv import load_dotenv
load_dotenv(".env")

URI = os.getenv("NEO4J_URI")
USER = os.getenv("NEO4J_USER")
PASSWORD = os.getenv("NEO4J_PASSWORD")

driver = GraphDatabase.driver(URI, auth=(USER, PASSWORD))

NUM_USERS = 1000
NUM_TRACKS = 5000

ALPHA_USER_ACTIVITY = 2.2   #How many TRACKS a USER LISTENS to
ALPHA_TRACK_POP = 2.0       #Popularity of TRACKS
ALPHA_SOCIAL = 2.3          #Social Connectivity

def create_users_batched(session):
    users = [{"id" : f"user_{i}", "name" : f"User {i}"} for i in range(NUM_USERS)]
    session.execute_write(lambda tx : tx.run("""
                                             UNWIND $batch AS user
                                             CREATE (:User {user_id: user.id, name: user.name})
                                             """, batch = users))
    
def create_tracks_batched(session):
    tracks = [{"id" : f"track_{i}", "title" : f"Track {i}"} for i in range(NUM_TRACKS)]
    session.execute_write(lambda tx : tx.run("""
                                             UNWIND $batch AS track
                                             CREATE (:Track {track_id: track.id, title: track.title})
                                             """, batch = tracks))
    
def generate_listens(session):
    #Pre-calculate track popularity weights
    #This is so that the track_0 is a "GLOBAL_HIT" and track_4999 is "Niche"
    track_indices = np.arange(1, NUM_TRACKS + 1)
    weights = 1 / (track_indices ** ALPHA_TRACK_POP)
    weights /= weights.sum()
    track_ids = [f"track_{i}" for i in range(NUM_TRACKS)]
    
    for i in tqdm(range(NUM_USERS), desc="Listens"):
        uid = f"user_{i}"
        num_to_listen = min(np.random.zipf(ALPHA_USER_ACTIVITY), 200)
        
        #Weighted selection of tracks
        chosen = np.random.choice(track_ids, size = min(num_to_listen, NUM_TRACKS), replace = False, p = weights)
        
        #Create a batch
        batch = [{"tid": tid, "count" : int(min(np.random.zipf(1.5), 100))} for tid in chosen]
        
        session.execute_write(lambda tx : tx.run("""
                                                 MATCH (u : User {user_id: $uid})
                                                 UNWIND $batch AS item
                                                 MATCH (t : Track {track_id: item.tid})
                                                 CREATE (u)-[r:LISTENS]->(t)
                                                 SET r.count = item.count
                                                 """, uid = uid, batch = batch))
        

def generate_follows(session):
    user_ids = [f"user_{i}" for i in range(NUM_USERS)]
    
    for i in tqdm(range(NUM_USERS), desc = "Follows"):
        uid = f"user_{i}"
        num_follows = min(np.random.zipf(ALPHA_SOCIAL), 50)
        
        #Randomly sample friends
        targets = random.sample(user_ids, min(num_follows, NUM_USERS))
        batch = [tid for tid in targets if tid != uid]
        
        session.execute_write(lambda tx : tx.run("""
                                                 MATCH (u1 : User {user_id: $uid})
                                                 UNWIND $batch AS target_id
                                                 MATCH (u2: User {user_id: target_id})
                                                 CREATE (u1)-[:FOLLOWS]->(u2)
                                                 """, uid = uid, batch = batch))
        
with driver.session() as session:
    print("Creating Nodes....")
    create_users_batched(session)
    create_tracks_batched(session)
    
    print("Creating Relationships....")
    generate_listens(session)
    generate_follows(session)
    
    
driver.close()

print("Graph Database Populated")