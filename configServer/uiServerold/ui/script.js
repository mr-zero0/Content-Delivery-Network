const API_BASE_URL = window.location + "../";  // Backend API URL
 
document.addEventListener("DOMContentLoaded", function () {
    loadDSAndCN();  // Load data when the page is loaded
 
    // Event listeners for Delivery Services buttons
    document.getElementById("getDS").addEventListener("click", () => loadPage('getDeliveryService.html'));
    document.getElementById("addDS").addEventListener("click", () => loadPage('addDeliveryService.html'));
    document.getElementById("modDS").addEventListener("click", () => loadPage('modifyDeliveryService.html'));
    document.getElementById("delDS").addEventListener("click", () => loadPage('deleteDeliveryService.html'));
 
    // Event listeners for Config Nodes buttons
    document.getElementById("getCN").addEventListener("click", () => loadPage('getConfigNode.html'));
    document.getElementById("addCN").addEventListener("click", () => loadPage('addConfigNode.html'));
    document.getElementById("modCN").addEventListener("click", () => loadPage('modifyConfigNode.html'));
    document.getElementById("delCN").addEventListener("click", () => loadPage('deleteConfigNode.html'));

        // Event listener for Invalidate button
    document.getElementById("invalidate").addEventListener("click", invalidateCache);
});
 // Function to load a specific page
function loadPage(pageName) {
    window.location.href = pageName;
}
// Function to load Config Nodes and Delivery Services on the home page
function loadDSAndCN() {
    // Fetch Delivery Services
    let url = API_BASE_URL + "ds"
    let value = ""
    fetch(url)
        .then(response => {            
            value = response
            return response.json()
        })
        .then(data => {
            let table = "<table border='1'><tr><th></th><th>Name</th></tr>";
            data.serviceList.forEach(ds => {
                console.log(ds)
                table += `<tr><td><input type="radio" name="dsSelected" value="${ds}"</td><td>${ds}</td></tr>`;
            });
            table += "</table>";
            document.getElementById("dss").innerHTML = table;
        })
        .catch(error => {
            console.error("Error fetching delivery services:", error);
            console.error("data"+data)
            document.getElementById("dss").innerHTML = "<p>"+data+"</p>"+"<p>Error loading Delivery Services from "+url+"."+error+"</p>";
        });
 
    // Fetch Config Nodes
    url = API_BASE_URL + "cn"
    fetch(url)
        .then(response => response.json())
        .then(data => {
            console.log(data)
            let table = "<table border='1'><tr><th></th><th>Node Name</th></tr>";
            data.NodeList.forEach(cn => {
                table += `<tr>
                    <td><input type="radio" name="cnSelected" value="${cn}"></td>
                    <td>${cn}</td>
                </tr>`;
            });
            table += "</table>";
            document.getElementById("cns").innerHTML = table;
        })
        .catch(error => {
            console.error("Error fetching config nodes:", error);
            document.getElementById("cns").innerHTML = "<p>Error loading Config Nodes from "+url+"."+error+"</p>";
        });
}

function invalidateCache() {
    const pattern = document.getElementById("invalidatePattern").value;
    if (!pattern) {
        alert("Please enter a pattern to invalidate.");
        return;
    }

    const url = `${API_BASE_URL}invalidate/${pattern}`;
    fetch(url)
        .then(response => {
            if (response.ok) {
                return response.text();
            } else {
                throw new Error("Invalidation request failed.");
            }
        })
        .then(data => {
            document.getElementById("invalidateStatus").innerText = data;
        })
        .catch(error => {
            console.error("Error invalidating cache:", error);
            document.getElementById("invalidateStatus").innerText = `Error: ${error.message}`;
        });
}