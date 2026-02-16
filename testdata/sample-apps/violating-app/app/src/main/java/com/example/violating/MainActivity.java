package com.example.violating;

import android.app.Activity;
import android.os.Bundle;
import android.telephony.SmsManager;
import java.net.HttpURLConnection;
import java.net.URL;

public class MainActivity extends Activity {

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        // PDS001: Unencrypted data transmission - using HTTP instead of HTTPS
        sendDataOverHttp("http://api.example.com/user-data");

        // PDS002: Collecting advertising ID without proper disclosure
        collectAdvertisingId();

        // Send SMS without clear user consent
        sendSmsMessage("+1234567890", "Hello from violating app");
    }

    // PDS001: Cleartext HTTP traffic
    private void sendDataOverHttp(String urlString) {
        try {
            URL url = new URL("http://insecure-server.com/data");
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("POST");
            conn.getOutputStream().write("sensitive-data".getBytes());
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    // PDS002: Advertising ID collection
    private void collectAdvertisingId() {
        // Simulates getting advertising ID
        com.google.android.gms.ads.identifier.AdvertisingIdClient.getAdvertisingIdInfo(this);
    }

    // SDK001: Using outdated analytics SDK patterns
    private void initAnalytics() {
        // com.google.firebase.analytics.FirebaseAnalytics usage
        com.google.firebase.analytics.FirebaseAnalytics.getInstance(this);
    }

    // Sending SMS programmatically
    private void sendSmsMessage(String phone, String message) {
        SmsManager smsManager = SmsManager.getDefault();
        smsManager.sendTextMessage(phone, null, message, null, null);
    }

    // AD001: Account creation without deletion mechanism
    public void createAccount(String email, String password) {
        // Creates user account but no deletion flow exists
        UserRepository.createUser(email, password);
    }
}
