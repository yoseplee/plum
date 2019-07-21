package playground.rmi;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

public class Client {

    private Client() {}

    public static void main(String[] args) {

        String host = (args.length < 1) ? null : args[0];
        // int port = (args.length < 1) ? null : Integer.parseInt(args[1]);
        try {
            // System.out.println("HOST:: " + host + "port:: " + port);
            // Registry registry = LocateRegistry.getRegistry(host);
            Registry registry = LocateRegistry.getRegistry("127.0.0.1:32120");
            Hello stub = (Hello) registry.lookup("Hello");
            // Hello stub = (Hello) registry.lookup("Hello1");
            // Hello stub = (Hello) registry.lookup("Hello2");
            String response = stub.sayHello();
            System.out.println("response: " + response);
        } catch (Exception e) {
            System.err.println("Client exception: " + e.toString());
            e.printStackTrace();
        }
    }
}
