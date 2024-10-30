function getAccessTokenCookie() {
    let value = `; ${document.cookie}`;
    let parts = value.split(`; access-token=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}

function deleteAccessTokenCookie() {
    document.cookie = 'access-token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
}