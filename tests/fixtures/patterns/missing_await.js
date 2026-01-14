// Missing await - should be detected
async function fetchData() {
    return "data";
}

async function processData() {
    fetchData(); // Missing await
    const result = fetchData(); // Missing await
    return result;
}












