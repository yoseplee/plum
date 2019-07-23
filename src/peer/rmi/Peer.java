package peer.rmi;

import java.rmi.registry.Registry;
import java.rmi.registry.LocateRegistry;
import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
import java.util.ArrayList;

public class Peer implements MessageIF {

    private long idx;
    private ArrayList<String> addressBook;

    public Peer() {}

    @Override
    public String sayHello() throws RemoteException {
        return "Hi I'm Peer";
    }

    @Override
    public void sendMessage(String message) throws RemoteException {
        System.out.println("I've got message:: " + message);
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