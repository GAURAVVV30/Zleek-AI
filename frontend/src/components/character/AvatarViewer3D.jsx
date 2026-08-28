import React, { useEffect, useRef } from 'react';
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';

export default function AvatarViewer3D({ character, customColor = null, auraColor = '#06b6d4' }) {
  const mountRef = useRef(null);
  const modelRef = useRef(null);
  const originalColorsRef = useRef(new Map());

  useEffect(() => {
    if (!mountRef.current || !character) return;
    const currentMount = mountRef.current;

    // 1. Scene Setup
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(
      45,
      currentMount.clientWidth / currentMount.clientHeight,
      0.1,
      1000
    );
    camera.position.set(0, 1.3, 3.2);

    // 2. High-Fidelity WebGL Renderer
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(currentMount.clientWidth, currentMount.clientHeight);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    renderer.toneMappingExposure = 1.3;
    renderer.outputColorSpace = THREE.SRGBColorSpace;
    currentMount.innerHTML = '';
    currentMount.appendChild(renderer.domElement);

    // 3. Orbit Controls
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.05;
    controls.target.set(0, 0.8, 0);
    controls.enablePan = false;
    controls.enableZoom = true;
    controls.minDistance = 2.0;
    controls.maxDistance = 4.5;
    controls.minPolarAngle = Math.PI / 3;
    controls.maxPolarAngle = Math.PI / 1.8;
    controls.autoRotate = true;
    controls.autoRotateSpeed = 1.2;

    // 4. Free-Fire / Cinematic Dynamic Lighting
    const ambientLight = new THREE.AmbientLight(0xffffff, 1.2);
    scene.add(ambientLight);

    // Key Front Light
    const keyLight = new THREE.DirectionalLight(0xffffff, 2.5);
    keyLight.position.set(4, 5, 4);
    scene.add(keyLight);

    // Cyber Rim Light 1 (Cyan / Blue edge glow)
    const rimLight1 = new THREE.DirectionalLight(new THREE.Color(auraColor), 3.5);
    rimLight1.position.set(-5, 3, -4);
    scene.add(rimLight1);

    // Warm Rim Light 2 (Orange / Magenta highlight)
    const rimLight2 = new THREE.DirectionalLight(0xff007f, 2.0);
    rimLight2.position.set(4, -2, -3);
    scene.add(rimLight2);

    // 5. Hologram Cyber Platform Ring (Free Fire style battle podium)
    const ringGeometry = new THREE.RingGeometry(1.0, 1.08, 64);
    const ringMaterial = new THREE.MeshBasicMaterial({
      color: new THREE.Color(auraColor),
      side: THREE.DoubleSide,
      transparent: true,
      opacity: 0.85,
    });
    const platformRing = new THREE.Mesh(ringGeometry, ringMaterial);
    platformRing.rotation.x = -Math.PI / 2;
    platformRing.position.y = 0.02;
    scene.add(platformRing);

    const outerRingGeo = new THREE.RingGeometry(1.3, 1.33, 64);
    const outerRingMat = new THREE.MeshBasicMaterial({
      color: 0xffffff,
      side: THREE.DoubleSide,
      transparent: true,
      opacity: 0.35,
    });
    const outerRing = new THREE.Mesh(outerRingGeo, outerRingMat);
    outerRing.rotation.x = -Math.PI / 2;
    outerRing.position.y = 0.01;
    scene.add(outerRing);

    // 6. Floating Ambient Particle Sparkles
    const particleCount = 120;
    const particleGeometry = new THREE.BufferGeometry();
    const positions = new Float32Array(particleCount * 3);

    for (let i = 0; i < particleCount * 3; i += 3) {
      positions[i] = (Math.random() - 0.5) * 4;
      positions[i + 1] = Math.random() * 2.5;
      positions[i + 2] = (Math.random() - 0.5) * 4;
    }
    particleGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));

    const particleMaterial = new THREE.PointsMaterial({
      color: new THREE.Color(auraColor),
      size: 0.035,
      transparent: true,
      opacity: 0.7,
      blending: THREE.AdditiveBlending,
    });
    const particles = new THREE.Points(particleGeometry, particleMaterial);
    scene.add(particles);

    // 7. GLTF 3D Model Loading & Custom Material Tinting
    const primaryMaterialNames = ['Main', 'Body', 'Primary', 'Material_MR', 'Character', 'Robot', 'Skin', 'Armor'];

    const applyCustomColor = (model, color) => {
      model.traverse((object) => {
        if (object instanceof THREE.Mesh && object.material) {
          const material = object.material;
          const originalColor = originalColorsRef.current.get(material.uuid);
          if (!originalColor) return;

          if (primaryMaterialNames.some((n) => material.name.toLowerCase().includes(n.toLowerCase())) || originalColorsRef.current.size < 5) {
            material.color.set(color ? new THREE.Color(color) : originalColor);
            if (color) {
              material.roughness = 0.25;
              material.metalness = 0.6;
            }
          }
        }
      });
    };

    if (modelRef.current) scene.remove(modelRef.current);
    originalColorsRef.current.clear();

    const loader = new GLTFLoader();
    loader.load(
      character.url,
      (gltf) => {
        const currentModel = gltf.scene;

        const box = new THREE.Box3().setFromObject(currentModel);
        const size = new THREE.Vector3();
        box.getSize(size);
        const center = box.getCenter(new THREE.Vector3());

        const targetHeight = 1.6;
        const scaleFactor = size.y > 0 ? targetHeight / size.y : 1;
        currentModel.scale.setScalar(scaleFactor);

        currentModel.position.x -= center.x * scaleFactor;
        currentModel.position.y -= box.min.y * scaleFactor;
        currentModel.position.z -= center.z * scaleFactor;

        currentModel.traverse((object) => {
          if (object instanceof THREE.Mesh && object.material) {
            originalColorsRef.current.set(
              object.material.uuid,
              object.material.color ? object.material.color.clone() : new THREE.Color(0xffffff)
            );
          }
        });

        modelRef.current = currentModel;
        scene.add(currentModel);
        applyCustomColor(currentModel, customColor);
      },
      undefined,
      (err) => console.error('Error loading 3D character:', err)
    );

    // 8. Animation Render Loop
    let clock = new THREE.Clock();
    const animate = () => {
      const elapsedTime = clock.getElapsedTime();

      // Pulse hologram podium rings
      platformRing.rotation.z = elapsedTime * 0.5;
      outerRing.rotation.z = -elapsedTime * 0.3;

      // Float particles gently
      particles.position.y = Math.sin(elapsedTime * 0.8) * 0.08;

      controls.update();
      renderer.render(scene, camera);
    };
    renderer.setAnimationLoop(animate);

    // 9. Resize Handling
    const handleResize = () => {
      if (!currentMount) return;
      camera.aspect = currentMount.clientWidth / currentMount.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(currentMount.clientWidth, currentMount.clientHeight);
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      renderer.setAnimationLoop(null);
      if (currentMount && renderer.domElement && currentMount.contains(renderer.domElement)) {
        currentMount.removeChild(renderer.domElement);
      }
      renderer.dispose();
    };
  }, [character, customColor, auraColor]);

  return (
    <div className="relative w-full h-full min-h-[380px] sm:min-h-[460px] flex items-center justify-center select-none cursor-grab active:cursor-grabbing">
      {/* 3D WebGL Canvas */}
      <div ref={mountRef} className="w-full h-full absolute inset-0 z-10" />

      {/* Orbit Tip Hint */}
      <div className="absolute bottom-4 left-1/2 -translate-x-1/2 z-20 px-3 py-1 bg-slate-900/60 backdrop-blur-md rounded-full border border-white/20 text-[10px] text-white/80 font-mono flex items-center gap-1.5 pointer-events-none">
        <span>🔄 Drag to 360° Rotate · Scroll to Zoom</span>
      </div>
    </div>
  );
}
