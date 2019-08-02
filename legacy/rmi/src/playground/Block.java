package playground;

public class Block {
    
    private long version;
    private String blockCreator; //Type can be modified
    private String merkleRoot;
    private MerkleTree blockBody;
    private String previousBlockHash;

    public Block() {
        
    }

    /* get message list as parameter, considering ArrayList or array */
    public void BuildTree() {
        MerkleTree tmpMerkleTree = new MerkleTree();

        //assign core information into this block from merkleTree
        this.merkleRoot = tmpMerkleTree.getMerkleRoot();
        this.blockBody = tmpMerkleTree;
    }
}