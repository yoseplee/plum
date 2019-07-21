package playground.rmi;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

public class Client {

    private Client() {}

    public static void main(String[] args) {

        String port = (args.length < 1) ? null : args[0];
        try {
            Registry registry = LocateRegistry.getRegistry(Integer.parseInt(port));
            Hello stub = (Hello) registry.lookup("Hello1");
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
