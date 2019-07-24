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

    public Peer() {
        this.addressBook = new ArrayList<String>();
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