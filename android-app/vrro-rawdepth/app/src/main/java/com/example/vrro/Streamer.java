package com.example.vrro;

import android.util.Log;

import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.SocketException;
import java.net.UnknownHostException;
import java.nio.FloatBuffer;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;


public class Streamer {
    private static final String TAG = Streamer.class.getSimpleName();
    private static InetAddress streamingAddress;
    private static DatagramSocket streamingSocket;
    private static ExecutorService executorService;
    public static void prepare() {
        try {
            executorService = Executors.newFixedThreadPool(1);
            streamingSocket = new DatagramSocket();
            streamingAddress = InetAddress.getByName("");
        } catch (SocketException | UnknownHostException e) {
            throw new RuntimeException(e);
        }
    }

    public static void close() {
        streamingSocket.close();
    }

    public static void stream(float[] cameraPose, FloatBuffer points) {
        executorService.execute(() -> {
            //Stream pointcloud to server
            if (points.hasArray() && points.array().length >= 4 && cameraPose.length == 3){
                boolean addComma = false;
                float[] floatArrayPoints = points.array();
                DatagramPacket packet;
                StringBuilder udpMessageBuffer = new StringBuilder("{");
                udpMessageBuffer.append("\"camera\":{");
                udpMessageBuffer.append("\"x\":").append(cameraPose[0]);
                udpMessageBuffer.append(",\"y\":").append(cameraPose[1]);
                udpMessageBuffer.append(",\"z\":").append(cameraPose[2]);
                udpMessageBuffer.append("},\"points\":[");
                Log.d(TAG, "point count: " + points.array().length);
                for (int i = 0; i < floatArrayPoints.length-3; i+=4) {
                    Log.d(TAG, "stream: " + (i) + "len: " + floatArrayPoints.length);
                    if (floatArrayPoints[i+3] < 0.9){
                        //-----> this is the issue <-------
                        continue;
                    }
                    if (addComma) {
                        udpMessageBuffer.append(",");
                    }
                    udpMessageBuffer.append("{");
                    udpMessageBuffer.append("\"x\":").append(floatArrayPoints[i]);
                    udpMessageBuffer.append(",\"y\":").append(floatArrayPoints[i+1]);
                    udpMessageBuffer.append(",\"z\":").append(floatArrayPoints[i+2]);
                    udpMessageBuffer.append(",\"c\":").append(floatArrayPoints[i+3]);
                    udpMessageBuffer.append("}");
                    addComma = true;
                    byte[] streamingBuffer = udpMessageBuffer.toString().getBytes();
                    packet = new DatagramPacket(streamingBuffer, streamingBuffer.length, streamingAddress, 8080);
                    try {
                        streamingSocket.send(packet);
                        udpMessageBuffer.delete(0, udpMessageBuffer.length());
                    } catch (IOException e) {
                        throw new RuntimeException(e);
                    }
                }
                udpMessageBuffer.append("]}");
                byte[] streamingBuffer = udpMessageBuffer.toString().getBytes();
                packet = new DatagramPacket(streamingBuffer, streamingBuffer.length, streamingAddress, 8080);
                try {
                    streamingSocket.send(packet);
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
            }
        });
    }
}
