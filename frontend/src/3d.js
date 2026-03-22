import * as THREE from "three";
import Stats from "three/addons/libs/stats.module.js"

const container = document.getElementById("canvas-container");

// 1. Scene Setup
const scene = new THREE.Scene();
const camera = new THREE.PerspectiveCamera(
    75,
    container.clientWidth / container.clientHeight,
    0.1,
    1000,
);
camera.position.z = 5;

const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setClearColor(0x000000, 0);
renderer.setSize(container.clientWidth, container.clientHeight);
container.appendChild(renderer.domElement);
renderer.setPixelRatio(window.devicePixelRatio)

// Point Cloud Data
const count = 3500;
const positions = new Float32Array(count * 3);
const orbitData = []; // store custom orbit info

for (let i = 0; i < count; i++) {
    // const radius = Math.random() * 3 + 1;
    const radius = Math.random() * 3 + 0.3;
    const theta = Math.random() * Math.PI * 2; // latitude
    const phi = Math.random() * Math.PI * 2; // longitude

    orbitData.push({
        radius,
        theta,
        phi,
        speed: 0.001,
    });
}

const geometry = new THREE.BufferGeometry;
const material = new THREE.PointsMaterial({
    color: 0x00ffcc,
    size: 0.03,
    transparent: true,
    opacity: 0.6,
    map: new THREE.CanvasTexture((() => {
            const c = document.createElement('canvas'); c.width = 64; c.height = 64;
            const ctx = c.getContext('2d');
            ctx.beginPath(); ctx.arc(32, 32, 30, 0, Math.PI * 2); ctx.fillStyle = '#fff'; ctx.fill();
            return c;
        })()),
});

const points = new THREE.Points(geometry, material);
scene.add(points);

// const stats = new Stats();
// stats.showPanel(0); // 0: fps, 1: ms, 2: mb, 3+: custom


// stats.dom.style.position = 'absolute';
// stats.dom.style.top = '0px';
// stats.dom.style.right = '0px';
// stats.dom.style.left = 'auto';
// stats.dom.style.transform = 'scale(2.0)';
// stats.dom.style.transformOrigin = 'top right';

// document.body.appendChild(stats.dom);

// Animation Loop
function animate() {
    requestAnimationFrame(animate);

    // stats.begin();

    const positionsAttr = [];

    orbitData.forEach((p) => {
        p.theta += p.speed;
        p.phi += p.speed * 0.5;

        const x = p.radius * Math.sin(p.phi) * Math.cos(p.theta);
        const y = p.radius * Math.sin(p.phi) * Math.sin(p.theta);
        const z = p.radius * Math.cos(p.phi);

        positionsAttr.push(x, y, z);
    });

    geometry.setAttribute("position", new THREE.Float32BufferAttribute(positionsAttr, 3));

    // Subtle rotation of the whole cloud
    points.rotation.y += 0.002;

    renderer.render(scene, camera);

    // stats.end();
}

// 4. Handle Resizing
window.addEventListener("resize", () => {
    const width = container.clientWidth;
    const height = container.clientHeight;
    renderer.setSize(width, height);
    camera.aspect = width / height;
    camera.updateProjectionMatrix();
});



// // 2. Update in your animation loop
// function animate() {
//     stats.begin();

//     // Your rendering logic
//     renderer.render(scene, camera);

//     stats.end();
//     requestAnimationFrame(animate);
// }

animate();
