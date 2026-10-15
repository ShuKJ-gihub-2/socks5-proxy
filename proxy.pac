function FindProxyForURL(url, host) {
    // HTTPSのあとに半角スペースを1つ空けて、アドレスをそのまま書きます
    return "HTTPS soocks5-proxy.onrender.com:443";
}
