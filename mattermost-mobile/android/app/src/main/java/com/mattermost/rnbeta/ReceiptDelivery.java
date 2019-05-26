package com.mattermost.rnbeta;

import android.content.Context;
import android.support.annotation.Nullable;
import android.util.Log;
import java.lang.System;

import okhttp3.Call;
import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

import org.json.JSONObject;
import org.json.JSONException;

import com.mattermost.react_native_interface.ResolvePromise;
import com.facebook.react.bridge.ReactApplicationContext;
import com.facebook.react.bridge.WritableMap;
import com.oblador.keychain.KeychainModule;

public class ReceiptDelivery {
    public static void send (Context context, final String ackId, final String type) {
        final ReactApplicationContext reactApplicationContext = new ReactApplicationContext(context);
        final KeychainModule keychainModule = new KeychainModule(reactApplicationContext);

        keychainModule.getGenericPasswordForOptions(null, new ResolvePromise() {
            @Override
            public void resolve(@Nullable Object value) {
                if (value instanceof Boolean && !(Boolean)value) {
                    return;
                }

                WritableMap map = (WritableMap) value;
                if (map != null) {
                    String[] credentials = map.getString("password").split(",[ ]*");
                    String token = null;
                    String serverUrl = null;
                    if (credentials.length == 2) {
                        token = credentials[0];
                        serverUrl = credentials[1];

                    }

                    Log.i("ReactNative", String.format("Send receipt delivery ACK=%s TYPE=%s to URL=%s with TOKEN=%s", ackId, type, serverUrl, token));
                    execute(serverUrl, token, ackId, type);
                }
            }
        });
    }

    protected static void execute(String serverUrl, String token, String ackId, String type) {
        if (token == null || serverUrl == null) {
            return;
        }

        JSONObject json;
        long receivedAt = System.currentTimeMillis();

        try {
            json = new JSONObject();
            json.put("id", ackId);
            json.put("received_at", receivedAt);
            json.put("platform", "android");
            json.put("type", type);
        } catch (JSONException e) {
            Log.e("ReactNative", "Receipt delivery failed to build json payload");
            return;
        }

        final OkHttpClient client = new OkHttpClient();
        final MediaType JSON = MediaType.parse("application/json; charset=utf-8");
        RequestBody body = RequestBody.create(JSON, json.toString());
        Request request = new Request.Builder()
                .header("Authorization", String.format("Bearer %s", token))
                .header("Content-Type", "application/json")
                .url(String.format("%s/api/v4/notifications/ack", serverUrl.replaceAll("/$", "")))
                .post(body)
                .build();

        try {
            client.newCall(request).execute();
        } catch (Exception e) {
            Log.e("ReactNative", "Receipt delivery failed to send");
        }

    }
}
