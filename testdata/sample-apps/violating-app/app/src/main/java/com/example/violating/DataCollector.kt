package com.example.violating

import android.content.Context
import android.location.LocationManager
import android.Manifest
import android.content.pm.PackageManager

class DataCollector(private val context: Context) {

    // PDS003: Location data collection without privacy policy
    fun collectLocationData() {
        val locationManager = context.getSystemService(Context.LOCATION_SERVICE) as LocationManager
        val location = locationManager.getLastKnownLocation(LocationManager.GPS_PROVIDER)
        // Sends location to server without user consent
        sendToServer("http://tracker.example.com/location", location.toString())
    }

    // PDS001: More cleartext HTTP usage
    fun sendToServer(url: String, data: String) {
        val connection = java.net.URL(url).openConnection() as java.net.HttpURLConnection
        connection.requestMethod = "POST"
        connection.outputStream.write(data.toByteArray())
    }

    // PDS004: Third-party SDK data sharing without disclosure
    fun initThirdPartySDKs() {
        // Facebook SDK initialization - requires data safety disclosure
        com.facebook.FacebookSdk.sdkInitialize(context)
        // Google Ads SDK
        com.google.android.gms.ads.MobileAds.initialize(context)
    }

    // No privacy policy URL anywhere in the app
    // Missing user consent for data collection
}
