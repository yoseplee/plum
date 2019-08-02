package peer.rmi;

import java.rmi.AccessException;
import java.rmi.NotBoundException;
import java.rmi.RemoteException;
import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;
import java.util.ArrayList;

public class Client implements Runnable {

    private Registry registry;
    private MessageIF stub;
    private String host;
    private int port;
    private ArrayList<String> addressbook;

    //constructor overloading
    public Client() {
        this("127.0.0.1", 1099);
        this.host = "127.0.0.1"; 
        this.port = 1099;
    }
    public Client(String host) {
        this(host, 1099);
    }

    public Client(String host, int port) {
        this.host = host;
        this.port = port;
        addressbook = new ArrayList<String>();
    }   

    @Override
    public void run() {

    }

    //define remote call methods
    public String sayHello() {
        String result = "";
        try {    
            String response = stub.sayHello();
            stub.sendMessage("Hello peer!");
            result = response;
        } catch (Exception e) {
            System.err.println("Error! Did you connect to a peer?");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }
        return result;
    }

    public String getIP() {
        String result = "";
        try {    
            String response = stub.getIP();
            result = response;
        } catch (Exception e) {
            System.err.println("Error! Did you connect to a peer?");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }
        return result;
    }

    public String addAddress(String address) {
        String result = "";
        try {    
            String response = stub.addAddress(address);
            result = response;
        } catch (Exception e) {
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }
        return result;
    }

    public String addAddressbook() {
        //add addressbook to the peer
        String result = "";
        try {    
            System.out.println("send and have connected peer set addressbook");
            String response = stub.setAddressBook(addressbook);
            result = response;
        } catch (Exception e) {
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }
        return result;
    }

    public String printAddressbook() {
        String result = "";
        try {    
            System.out.println("addressbook of connected peer: " + this.host);
            String response = stub.printAllAddressBook();
            result = response;
        } catch (Exception e) {
            System.err.println("Error! Did you connect to a peer?");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }
        return result;
    }

    public String addTransaction(String transaction) {
        String result = "";
        try {    
            System.out.println("add transaction to connected peer: " + this.host);
            String response = stub.addTransaction(transaction);
            result = response;
        } catch (Exception e) {
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }
        return result;
    }

    public String printMempool() {
        String result = "";
        try {    
            System.out.println("mempool of connected peer: " + this.host);
            String response = stub.printMempool();
            result = response;
        } catch (Exception e) {
            System.err.println("Error! Did you connect to a peer?");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }
        return result;
    }

    @Override
    public String toString() {
        String result = "";
        result += "======================\n";
        result += "< Client >\n";
        result += String.format("%-16s%-10s\n", host, port);
        result += "======================\n";
        return result;
    }

    public boolean connect() {
        try {
            registry = LocateRegistry.getRegistry(host);
        } catch (Exception e) {        
            System.err.println("Connection Exception! Check out host address");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
            return false;
        }
        try {
            stub = (MessageIF) registry.lookup("peer");
        } catch (AccessException e) {
            System.err.println("Connection Exception! Check out access priv");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
            return false;
        } catch (RemoteException e) {
            System.err.println("Connection Exception! Check out is remote available");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
            return false;
        } catch (NotBoundException e) {
            System.err.println("Connection Exception! Check out bound name is correct");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
            return false;
        }
        return true;
    }

    //getters and setters
    public void setRegistry(Registry registry) {
        this.registry = registry;
    }

    public void setStub(MessageIF stub) {
        this.stub = stub;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public void setPort(int port) {
        this.port = port;
    }

    public void setAddressbook(ArrayList<String> addressbook) {
        this.addressbook = addressbook;
    }

    public Registry getRegistry() {
        return this.registry;
    }

    public MessageIF getStub() {
        return this.stub;
    }

    public String getHost() {
        return this.host;
    }

    public int getPort() {
        return this.port;
    }

    public ArrayList<String> getAddressbook() {
        return this.addressbook;
    }
}