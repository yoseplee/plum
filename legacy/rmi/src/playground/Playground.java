package playground;

public class Playground {

    public static void main(String[] args) {
        System.out.println("THIS IS A PLAYGROUND FOR BLOCKCHAIN PEER PROGRAMMING");
        //test for a peer
        Peer peer1 = new Peer();
    
        //make message list to send
        for(int i=0; i<15; i++) {
            String messageToSend = (i+1)+"";
            System.out.println(messageToSend);
        }


        //peer1 create a block

        //check block
    }

    public void sendMessage(Peer peer, String message) {
        //send a message to specific peer

        //that peer verify message, and then gossip it on peer side
    }
}