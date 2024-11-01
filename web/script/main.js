function getAccessTokenCookie() {
    let value = `; ${document.cookie}`
    let parts = value.split(`; access-token=`)
    if (parts.length === 2) return parts.pop().split(';').shift()
}

function deleteAccessTokenCookie() {
    document.cookie = 'access-token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;'
}

async function uploadCSV(url, token, lastEditionBy, sectionId) {
    console.log('lasteditionby: ' + lastEditionBy, 'sectionid: ' + sectionId)
    const fileInput = document.getElementById("fileInput")
    if (fileInput.files.length === 0) {
        console.log('No file selected.')
    } else {
        const formData = new FormData();
        formData.append("file", fileInput.files[0]);

        formData.append("lastEditionBy", lastEditionBy);
        formData.append("sectionId", sectionId);
        try {
            const response = await fetch(url + '/csv', {
                method: 'POST',
                body: formData,
                headers: {
                    "Authorization": "Bearer " + token
                }
            })
            let body = await response.json()
            if (response.ok) {
                console.log('Resource ' + JSON.stringify(body.resource))
                sessionStorage.setItem('resource', JSON.stringify(body.resource))
            } else {
                console.error("File upload failed:", response.status, response.statusText, body)
            }
            console.log('File uploaded successfully!')
        } catch (error) {
            console.error("Error uploading file:", error)
        }
    }
}