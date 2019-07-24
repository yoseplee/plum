package peer.rmi;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

public class Client {

    private Registry registry;
    private MessageIF stub;
    private String host;

    public Client() {}
    public Client(String host) {
        this.host = host;
    }

    public void connect(String host) {
        try {
            registry = LocateRegistry.getRegistry(host);
            stub = (MessageIF) registry.lookup("peer");
        } catch (Exception e) {
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }        
    }

    public MessageIF getStub() {
        return this.stub;
    }

    public static void main(String[] args) {

        String host = (args.length < 1) ? null : args[0];
        Client client = new Client(host);
        client.connect(host);
        MessageIF stubFromClient = client.getStub();

        try {    
            String response = stubFromClient.sayHello();
            stubFromClient.sendMessage("Hello peer!");
            System.out.println("response: " + response);

            response = stubFromClient.getIP();
            System.out.println("response: " + response);
        } catch (Exception e) {
            
        }
    }
}
