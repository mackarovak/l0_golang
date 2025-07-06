async function getOrder() {
    const orderId = document.getElementById('orderId').value.trim();
    if (!orderId) {
        alert('Please enter Order ID');
        return;
    }

    try {
        const response = await fetch(`http://localhost:8080/order/${orderId}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        document.getElementById('result').innerHTML = 
            `<pre>${JSON.stringify(data, null, 2)}</pre>`;
    } catch (error) {
        document.getElementById('result').innerHTML = 
            `<p style="color: red;">Error: ${error.message}</p>`;
    }
}