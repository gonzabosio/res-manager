function getAccessTokenCookie() {
    let value = `; ${document.cookie}`;
    let parts = value.split(`; access-token=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}