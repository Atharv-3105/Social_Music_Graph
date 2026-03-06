import torch
import torch.nn.functional as F 
from torch_geometric.data import Data
from torch_geometric.utils import negative_sampling
from sklearn.model_selection import train_test_split
from tqdm import tqdm 
from model import GraphSAGE

#Load The Graph
data_dict = torch.load("graph_data2.pt")

x = data_dict["x"]
edge_index = data_dict["edge_index"]

num_nodes = x.shape[0]
data = Data(x = x, edge_index= edge_index)


#Split edges for training and testing
edge_index = data.edge_index
num_edges = edge_index.shape[1]

edge_list = edge_index.t()

train_edges, test_edges = train_test_split(
    edge_list,
    test_size=0.2,
    random_state=42
)

train_edges = train_edges.t().contiguous()
test_edges = test_edges.t().contiguous()

    
#Initialize the Model    
model = GraphSAGE(in_channels = x.shape[1], hidden_channels=64, out_channels=32)

optimizer = torch.optim.Adam(model.parameters(), lr=0.01)


def train():
    model.train()
    optimizer.zero_grad()
    
    z = model(data.x, train_edges)
    
    #Positive Edges
    pos_edge_idx = train_edges
    
    #Negative Sampling
    neg_edge_index = negative_sampling(edge_index=pos_edge_idx, 
                                       num_nodes=num_nodes,
                                       num_neg_samples=pos_edge_idx.shape[1])
    
    
    #Positive Scores
    pos_score = (z[pos_edge_idx[0]] * z[pos_edge_idx[1]]).sum(dim = 1)
    
    #Negatice Scores
    neg_score = (z[neg_edge_index[0]] * z[neg_edge_index[1]]).sum(dim = 1)
    
    #Loss
    loss = (F.binary_cross_entropy_with_logits(pos_score, torch.ones_like(pos_score)) +
            F.binary_cross_entropy_with_logits(neg_score, torch.zeros_like(neg_score)))

    loss.backward()
    optimizer.step()
    
    return loss.item()


def evaluate():
    model.eval()
    with torch.no_grad():
        
        z = model(data.x, train_edges)
        
        pos_edge_idx = test_edges
        neg_edge_idx = negative_sampling(edge_index=train_edges,
                                         num_nodes=num_nodes,
                                         num_neg_samples=pos_edge_idx.shape[1])
        
        pos_score = torch.sigmoid((z[pos_edge_idx[0]] * z[pos_edge_idx[1]]).sum(dim = 1))
        
        neg_score = torch.sigmoid((z[neg_edge_idx[0]] * z[neg_edge_idx[1]]).sum(dim = 1))
        
        correct = ((pos_score > 0.5).sum() + (neg_score < 0.5).sum())
        
        total = len(pos_score) + len(neg_score)
        
        return correct / total
    
    
    
#====================Training Loop======================
for epoch in tqdm(range(1, 101)):
    loss = train()
    
    if epoch % 10 == 0:
        acc = evaluate()
        print(f"Epoch: [{epoch}], Loss:[{loss:.4f}], Accuracy:[{acc:.4f}]")
        
        
#====================Save Embeddings=====================
model.eval()
with torch.no_grad():
    embeddings = model(data.x, train_edges)
    
torch.save(embeddings, "user_embeddings.pt")

print("Training Complete, Embeddings Saved.")