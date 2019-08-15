package plum.cli;

import plum.*;
import java.util.Scanner;
import java.util.ArrayList;

public class PlumCli {

    public static void help() {
        System.out.println("CLI Function List");
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "help", "print all the commands"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "connect", "connect to specific peer. type ip address(v4)"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "list", "list up all the clients where connection have"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "use", "select a client connected with a peer"));;
        
    }

    public static void clientHelp() {
        System.out.println("PlumClient Function List");
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "sayHello", "send say hello to connect peer"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "getIP", "get IP address from connected peer, and add it to client's local addressBook"));
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "addAddress", "send a single address to connected peer so that it can add it to its addressBook"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "setAddressbook", "send a list of address to connected peer so that it can add it to its addressBook"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "printAddressbook", "print all the addressess connected peer have"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "addTransaction", "add a transaction to peer and gossip"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "printMempool", "print all the mempool of connected peer have"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "exit", "quit client"));;

    }

    public static void printAllClients(ArrayList<PlumClient> clientList) {
        if(clientList.size()==0) {
            System.err.println("PlumClient List Empty");
            return;
        }
        int i = 0;
        System.out.println(String.format("%-5s%-16s%-10s", "idx", "|ip", "|port"));
        for(PlumClient tmpClient : clientList) {
            System.out.println(String.format("%-5d%-16s%-10s", ++i, tmpClient.getHost(), tmpClient.getPort()));
        }
    }

    public static void welcomeCli() {
        System.out.println("===================");
        System.out.println("WELCOME TO PLUM CLI!");
        System.out.println("===================");
        System.out.println("help command for listing up commands that client have");
    }

    public static void welcomePeer() {
        System.out.println("*******************");
        System.out.println("WELCOME TO PLUM PEER!");
        System.out.println("*******************");
        System.out.println("help command for listing up commands that client have");
    }

    public PlumCli() {}
    public static void main(String[] args) {
        ArrayList<PlumClient> clientList = new ArrayList<PlumClient>();
        PlumClient currentClient = null;

        // String host = (args.length < 1) ? null : args[0];

        // PlumClient client;
        Scanner scan = new Scanner(System.in);
        ArrayList<IPAddress> addressBook = new ArrayList<IPAddress>();
        welcomeCli();
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
                    PlumClient client = new PlumClient(host, 50051);
                    addressBook.add(client.getIP());
                    clientList.add(client);
                    System.out.println("connected to " + host);
                    break;           
                case "list":
                    printAllClients(clientList);
                    break;
                case "use":
                    int clientIdx = -1;
                    printAllClients(clientList);
                    System.out.print("which client to use? (idx) &> ");
                    clientIdx = Integer.parseInt(scan.nextLine());
                    //check bound
                    if(clientIdx > clientList.size() || clientIdx < 0) {
                        System.err.println("client idx out of bound!");
                        continue;
                    }
                    currentClient = clientList.get(clientIdx-1);
                    System.out.println(currentClient);
                    welcomePeer();
                    while(true) {
                        String response = "";
                        System.out.printf("%10s | &> ", currentClient.getHost());
                        String message = scan.nextLine();
                        //send RMI call
                        switch(message) {
                            case "help":
                                clientHelp();
                                break;
                            case "sayHello":
                                // response = currentClient.sayHello();
                                currentClient.sayHello("cli");
                                break;
                            case "getIP":
                                response = currentClient.getIP().getAddress();
                                break;
                            case "addAddress":
                                String address = "";
                                System.out.println("which address to add? &> ");
                                address = scan.nextLine();
                                IPAddress addrToAdd = IPAddress.newBuilder().setAddress(address).setPort(50051).build();
                                // response = currentClient.addAddress(addrToAdd);
                                break;
                            case "setAddressbook":
                                //set client addressBook from cli
                                try {
                                    currentClient.setAddressBook(addressBook);                     
                                } catch (InterruptedException e) {
                                    e.printStackTrace();
                                }
                                //set peer addressBook from client
                                // response = currentClient.addAddressbook();
                                break;
                            case "printAddressbook":
                                ArrayList<IPAddress> book = currentClient.getAddressBook();
                                for(IPAddress temp : book) {
                                    System.out.println("address:: " + temp.getAddress() + ":" + temp.getPort());
                                }
                                break;
                            case "addTransaction":
                                String transaction = "";
                                System.out.println("which transaction to add? &> ");
                                currentClient.addTransaction(Transaction.newBuilder().setTransaction(transaction).build());
                                break;
                            case "printMempool":
                                ArrayList<Transaction> memPool = currentClient.getMemPool();
                                for(Transaction temp : memPool) {
                                    System.out.println("transaction from peer mempool: " + temp.getTransaction());
                                }
                                break;
                            default:
                                break;
                        }
                        if(message.equals("exit")) {
                            System.out.println("..back to CLI");
                            break;
                        }
                        //print
                        if(!response.isEmpty()) System.out.println("response: " + response);
                    }
                    welcomeCli();
                    break;
                case "exit":
                    System.out.println("Bye");
                    scan.close();
                    return;
                default:
                    break;
            }
        }
    }
}