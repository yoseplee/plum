package peer.rmi;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;
import java.util.Scanner;
import java.util.ArrayList;

public class Cli {

    public Cli() {}

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

    public static void printAllClients(ArrayList<Client> clientList) {
        if(clientList.size()==0) {
            System.err.println("Client List Empty");
            return;
        }
        int i = 0;
        for(Client tmpClient : clientList) {
            System.out.printf("%-5s%-16s%-10s%", "idx", "|ip", "|port\n");
            System.out.printf("%-5s%-16s%-10s%-30s\n", ++i, tmpClient.getHost(), tmpClient.getPort());
        }
    }

    public static void main(String[] args) {
        ArrayList<Client> clientList = new ArrayList<Client>();
        Client currentClient = null;

        // String host = (args.length < 1) ? null : args[0];

        // Client client;
        Scanner scan = new Scanner(System.in);
        ArrayList<String> addressBook = new ArrayList<String>();
        System.out.println("===================");
        System.out.println("WELCOME TO PLUM PEER!");
        System.out.println("===================");
        System.out.println("help command for listing up commands that client have");
        while(true) {
            System.out.print("&> ");
            String command = "";
            command = scan.nextLine();
            switch(command) {
                case "help":
                    help();
                    break;
                case "connect":
                    System.out.print("Which host(v4)? &> ");
                    String host = scan.nextLine();
                    System.out.println("try to connect to host: " + host);
                    Client client = new Client(host);
                    client.connect();
                    break;           
                case "list":
                    printAllClients(clientList);
                    break;
                case "use":
                    int clientIdx = -1;
                    printAllClients(clientList);
                    System.out.println("which client to use? (idx) &> ");
                    clientIdx = Integer.parseInt(scan.nextLine());
                    currentClient = clientList.get(clientIdx);
                    break;
                default:
                    break;
            }
            
            /*else if(command.equals("sayHello")) {
                try {    
                    String response = stubFromClient.sayHello();
                    stubFromClient.sendMessage("Hello peer!");
                    System.out.println("response: " + response);
                } catch (Exception e) {
                    System.err.println("Error! Did you connect to a peer?");
                    System.err.println("Client exception: " + e.toString());
                    e.printStackTrace();
                }
            } else if(command.equals("getIP")) {
                try {    
                }
                    String response = stubFromClient.getIP();
                    System.out.println("response: " + response);
                    addressBook.add(response);
                } catch (Exception e) {
                    System.err.println("Error! Did you connect to a peer?");
                    System.err.println("Client exception: " + e.toString());
                    e.printStackTrace();
                }
            } else if(command.equals("address")) {
                try {    
                    String address = "";
                    System.out.println("which address to add? &> ");
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
                    System.out.println("addressbook of connected peer? &> ");
                    String response = stubFromClient.printAllAddressBook();
                    System.out.println("response: " + response);
                } catch (Exception e) {
                    System.err.println("Error! Did you connect to a peer?");
                    System.err.println("Client exception: " + e.toString());
                    e.printStackTrace();
                }
            } else if(command.equals("exit")) {
                System.out.println("Bye");
                break;
            }*/
        }
    }
}