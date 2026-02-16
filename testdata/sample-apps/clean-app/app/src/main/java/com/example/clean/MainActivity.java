package com.example.clean;

import android.app.Activity;
import android.os.Bundle;
import javax.net.ssl.HttpsURLConnection;
import java.net.URL;

public class MainActivity extends Activity {

    private static final String PRIVACY_POLICY_URL = "https://example.com/privacy-policy";

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // Proper HTTPS connection
        fetchData("https://api.example.com/data");
    }

    // Secure HTTPS communication
    private void fetchData(String urlString) {
        try {
            URL url = new URL(urlString);
            HttpsURLConnection conn = (HttpsURLConnection) url.openConnection();
            conn.setRequestMethod("GET");
            conn.getInputStream();
        } catch (Exception e) {
            // Proper error handling
            android.util.Log.e("MainActivity", "Network error", e);
        }
    }

    // Account management with deletion support
    public void createAccount(String email, String password) {
        UserRepository.createUser(email, password);
    }

    public void deleteAccount(String userId) {
        UserRepository.deleteUser(userId);
    }

    public String getPrivacyPolicyUrl() {
        return PRIVACY_POLICY_URL;
    }
}
