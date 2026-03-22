import { Events } from "@wailsio/runtime";
import { GreetService } from "../bindings/github.com/mszalewicz/veles";

const heightElement = document.getElementById("label-height")
const widthElement = document.getElementById("label-width")
const positionElement = document.getElementById("label-center")

window.doGreet = async () => {
    try {
        resultElement.innerText = await GreetService.Greet(name);
    } catch (err) {
        console.error(err);
    }
};


Events.On('window-resized', (event) => {
    const { width, height } = event.data;

    heightElement.innerText = `${height}`
    widthElement.innerText = `${width}`
});

Events.On('window-repositioned', (event) => {
    const {x, y} = event.data;

    positionElement.innerText = `x: ${x} | y: ${y}`
});


