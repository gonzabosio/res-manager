async function uploadCSV(url, token, lastEditionBy, sectionId, userId) {
    const fileInput = document.getElementById("fileInput")
    if (fileInput.files.length === 0) {
        console.log('No file selected.')
    } else {
        const formData = new FormData();
        formData.append("file", fileInput.files[0]);

        formData.append("lastEditionBy", lastEditionBy);
        formData.append("sectionId", sectionId);
        formData.append("userId", userId)
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
                sessionStorage.setItem('resource', JSON.stringify(body.resource))
                sessionStorage.setItem('resource-id', body.resource_id.toString())
                console.log('CSV uploaded')
            } else {
                console.error("File upload failed:", response.status, response.statusText, body)
            }
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
