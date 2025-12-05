// Code after throw - should be detected (enhanced unreachable code)
function test() {
    throw new Error("Error");
    console.log("This is unreachable");
    return "never reached";
}

function test2() {
    if (error) {
        throw new Error("Error");
        console.log("Unreachable");
    }
}

