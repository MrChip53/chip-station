async function loadWasm(wasm, target) {
  if (WebAssembly && !WebAssembly.instantiateStreaming) { // polyfill
    WebAssembly.instantiateStreaming = async (resp, importObject) => {
      const source = await (await resp).arrayBuffer();
      return await WebAssembly.instantiate(source, importObject);
     };
  }  

  const go = new Go();

  go.argv = [wasm, target];

  const result = await WebAssembly.instantiateStreaming(fetch(wasm), go.importObject);
  go.run(result.instance);
}

function attachRomUploadListeners() {
  const handleFile = (file) => {
    const reader = new FileReader();
    reader.onload = function(e) {
      const arrayBuffer = e.target.result;
      const uint8Array = new Uint8Array(arrayBuffer);
      startRom(uint8Array);
      document.body.style.backgroundColor = '#000';
    };
    reader.readAsArrayBuffer(file);
  };

  const dragOverListener = (event) => {
     event.preventDefault();
     document.body.style.backgroundColor = 'green';
  };

  const dragLeaveListener = (event) => {
    event.preventDefault();
    document.body.style.backgroundColor = '#000';
  };

  const dropListener = (event) => {
    event.preventDefault();
    document.body.style.backgroundColor = '#00F';
    if (event.dataTransfer.files.length > 0) {
      const file = event.dataTransfer.files[0];
      handleFile(file);
    } else {
      const text = event.dataTransfer.getData('text');
      if (text.startsWith('http') || text.startsWith('data:')) {
        fetch(text)
          .then(response => response.arrayBuffer())
          .then(buffer => {
            const uint8Array = new Uint8Array(buffer);
            startRom(uint8Array);
          });
      } else {
        loadRomFromText(text);
      }
      document.body.style.backgroundColor = '#000';
    }
  };

  document.addEventListener('dragover', dragOverListener);
  document.addEventListener('dragleave', dragLeaveListener);
  document.addEventListener('drop', dropListener);

  return { dragOverListener, dragLeaveListener, dropListener };
}

function attachVisibilityListener() {
  let runningOnHide = false;

  const visibilityChangeListener = () => {
    if (document.hidden) {
      runningOnHide = !emulator.isPaused();
      if (runningOnHide) {
        emulator.pause();
      }
    } else {
      if (runningOnHide) {
        emulator.resume();
      }
    }
  };

  document.addEventListener('visibilitychange', visibilityChangeListener);

  return { visibilityChangeListener };
}

function downloadRom() {
  const rom = emulator.getRom();
  const blob = new Blob([rom], { type: 'application/octet-stream' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'rom.bin';
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
}

(async () => {
  await loadWasm("main.wasm", "chip8-ui");
  attachVisibilityListener();
  
  console.log('ChipStation Initialized');
})();
