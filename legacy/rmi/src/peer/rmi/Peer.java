package peer.rmi;

import java.rmi.registry.Registry;
import java.rmi.registry.LocateRegistry;
import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
import java.net.InetAddress;
import java.util.ArrayList;

public class Peer implements MessageIF {

    private long idx;
    private long version;
    private ArrayList<String> addressBook;
    private ArrayList<String> mempool;

    public Peer() {
        this.addressBook = new ArrayList<String>();
        this.mempool = new ArrayList<String>();
    }

    @Override
    public String sayHello() throws RemoteException {
        return "Hi I'm Peer";
    }

    @Override
    public void sendMessage(String message) throws RemoteException {
        System.out.println("I've got message:: " + message);
    }

    @Override
    public String getIP() throws RemoteException {
        String myip = "";
        try {
            InetAddress local = InetAddress.getLocalHost();
            myip = local.getHostAddress().toString();
        } catch (Exception e) {
            System.err.println(e.getMessage());
        }
        return myip;
    }

    @Override
    public String setAddressBook(ArrayList<String> addressBook) {
        String message = "{success: true}";
        this.addressBook = addressBook;
        return message;
    }

    @Override
    public String addAddress(String address) {
        //need to implement duplicated address before add
        String message = "{success: true}";
        this.addressBook.add(address);
        return message;
    }

    @Override
    public String printAllAddressBook() {
        String message = "{";
        message += "{success: true},";
        message += "address: [";
        int size = addressBook.size();
        for(int i=0; i<size; i++) {
            String address = addressBook.get(i);
            message += address;
            if(i+1 != size) message += ", ";
        }
        message += "]}";

        return message;
    }

    @Override
    public String addTransaction(String transaction) {
        String message = "";
        message += "{success: true}, ";

        //conditions
        //verify transaction
        //if(verifyTransaction(message))
        //duplication check
        if(mempool.contains(transaction)) {
            return "{success: false, desc: duplicated}";
        }

        //and then add this transaction to mempool
        mempool.add(transaction);
        gossip(transaction);

        return message;
    }

    @Override
    public String printMempool() {
        String message = "{";
        message += "{success: true},";
        message += "transaction: [";
        int size = mempool.size();
        for(int i=0; i<size; i++) {
            String transaction = mempool.get(i);
            message += transaction;
            if(i+1 != size) message += ", ";
        }
        message += "]}";

        return message;
    }

    public void gossip(String gossipMessage) {
        //make a list of client to gossip something
        //basic thought is to fire and fotgot
        ArrayList<Client> clientList = new ArrayList<Client>();
        for(String host : addressBook) {
            Client client = new Client(host);
            client.connect();
            clientList.add(client);
        }

        //may a client act as thread
        String host = "000.000.000.000";
        try {
            host = getIP();
        } catch (RemoteException e) {
            System.err.println("GET IP ERROR: " + e.toString());
            e.printStackTrace();
        }
        System.out.println(String.format("%-10s%-5s%-2s%-5s%-2s%-30s", "GOSSIP:: ", "FROM", "-->", "TO", " | " , "gossipMessage"));
        for(Client tmpClient : clientList) {
            System.out.println(String.format("%-10s%-5s%-2s%-5s%-2s%-30s", "GOSSIP:: ", host, "-->", tmpClient.getHost(), " | " , gossipMessage));
            tmpClient.addTransaction(gossipMessage);
        }
    }

    public static void main(String args[]) {
        try {
            Peer obj = new Peer();
            MessageIF stub = (MessageIF) UnicastRemoteObject.exportObject(obj, 0);

            // bind the remote object's stub in the registry 
            Registry registry = LocateRegistry.getRegistry();
            registry.bind("peer", stub);

            System.err.println("Peer ready");
        } catch (Exception e) {
            System.err.println("Peer exception: " + e.toString());
            e.printStackTrace();
        }
    }
}