package peer.rmi;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;
import java.util.ArrayList;
import java.util.Scanner;

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

    public static void help() {
        System.out.println("Client Function List");
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "help", "print all the commands"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "connect", "connect to specific peer. type ip address(v4)"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "sayHello", "send say hello to connect peer"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "getIP", "get IP address from connected peer, and add it to client's local addressbook"));
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "address", "send a single address to connected peer so that it can add it to its addressbook"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "addressbook", "send a list of address to connected peer so that it can add it to its addressbook"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "print addressbook", "print all the addressess connected peer have"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "exit", "quit client"));;
    }

    public static void main(String[] args) {

        // String host = (args.length < 1) ? null : args[0];

        Client client;
        MessageIF stubFromClient = null;
        Scanner scan = new Scanner(System.in);
        ArrayList<String> addressBook = new ArrayList<String>();
        System.out.println("===================");
        System.out.println("WELCOME TO PEER");
        System.out.println("===================");
        System.out.println("help command for listing up commands that client have");
        while(true) {
            System.out.print("> ");
            String command = "";
            command = scan.nextLine();
            if(command.equals("help")) {
                help();
            } else if(command.equals("connect")) {
                System.out.print("Which host(v4)? > ");
                String host = scan.nextLine();
                System.out.println("try to connect to host: " + host);
                client = new Client(host);
                client.connect(host);
                stubFromClient = client.getStub();
            } else if(command.equals("sayHello")) {
                try {    
                    String response = stubFromClient.sayHello();
                    stubFromClient.sendMessage("Hello peer!");
                    System.out.println("response: " + response);
                } catch (Exception e) {
                    System.err.println("Client exception: " + e.toString());
                    e.printStackTrace();
                }
            } else if(command.equals("getIP")) {
                try {    
                    String response = stubFromClient.getIP();
                    System.out.println("response: " + response);
                    addressBook.add(response);
                } catch (Exception e) {
                    System.err.println("Client exception: " + e.toString());
                    e.printStackTrace();
                }
            } else if(command.equals("address")) {
                try {    
                    String address = "";
                    System.out.println("which address to add? > ");
                    address = scan.nextLine();
                    String response = stubFromClient.addAddress(address);
                    System.out.println("response: " + response);
                } catch (Exception e) {
                    System.err.println("Client exception: " + e.toString());
                    e.printStackTrace();
                }
            } else if(command.equals("addressbook")) {
                try {    
                    System.out.println("addressbook send");
                    String response = stubFromClient.setAddressBook(addressBook);
                    System.out.println("response: " + response);
                } catch (Exception e) {
                    System.err.println("Client exception: " + e.toString());
                    e.printStackTrace();
                }
            } else if(command.equals("print addressbook")) {
                try {    
                    System.out.println("addressbook of connected peer? > ");
                    String response = stubFromClient.printAllAddressBook();
                    System.out.println("response: " + response);
                } catch (Exception e) {
                    System.err.println("Client exception: " + e.toString());
                    e.printStackTrace();
                }
            } else if(command.equals("exit")) {
                System.out.println("Bye");
                break;
            }
        }
    }
}
