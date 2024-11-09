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
                console.log('Resource ' + body)
                sessionStorage.setItem('resource', body)
            } else {
                console.error("File upload failed:", response.status, response.statusText, body)
            }
            console.log('File uploaded successfully!')
        } catch (error) {
            console.error("Error uploading file:", error)
        }
    }
}

async function uploadImage(backURL, token, resourceId) {
    const fileInput = document.getElementById("imageFile")
    const file = fileInput.files[0]

    if (!file) {
        console.log("Please select a file to upload")
        return
    }

    const formData = new FormData()
    formData.append("image", file)
    formData.append("resourceId", resourceId);

    const endpoint = `${backURL}/image`
    const authToken = token;

    try {
        const response = await fetch(endpoint, {
            method: "POST",
            headers: {
                "Authorization": `Bearer ${authToken}`,
            },
            body: formData,
        })
        const resBody = await response.json()
        if (response.ok) {
            console.log("Image uploaded successfully", resBody)
            document.getElementById("info-message").innerText = "Image uploaded successfully"
        } else {
            console.log("Failed to upload image", resBody)
            document.getElementById("err-message").innerText = "Failed to upload image"
        }
    } catch (error) {
        console.error("Error uploading image:", error)
    }
}
