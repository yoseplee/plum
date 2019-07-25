package peer.rmi;

import java.rmi.AccessException;
import java.rmi.NotBoundException;
import java.rmi.RemoteException;
import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

public class Client {

    private Registry registry;
    private MessageIF stub;
    private String host;
    private int port;

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
    }   

    @Override
    public String toString() {
        String result = "";
        result += "======================\n";
        result += "< Client >\n";
        result += String.format("%-16s%-10s%-30s\n", host, port);
        result += "======================\n";
        return result;
    }

    public void connect() {
        try {
            registry = LocateRegistry.getRegistry(host);
            
        } catch (Exception e) {
            System.err.println("Connection Exception! Check out host address");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }
        try {
            stub = (MessageIF) registry.lookup("peer");
        } catch (AccessException e) {
            System.err.println("Connection Exception! Check out access priv");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        } catch (RemoteException e) {
            System.err.println("Connection Exception! Check out is remote available");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        } catch (NotBoundException e) {
            System.err.println("Connection Exception! Check out bound name is correct");
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
		}
    }

    //getters and setters
    public void setRegistry(Registry registry) {
        this.registry = registry;
    }

    public void setStub(MessageIF stub) {
        this.stub = stub;
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
}