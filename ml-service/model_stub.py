import numpy as np 

def get_user_embedding(user_id:str, dim = 64):
    seed = sum(ord(c) for c in user_id) % 2 ** 32
    rng = np.random.default_rng(seed)
    return rng.standard_normal(dim).tolist()
