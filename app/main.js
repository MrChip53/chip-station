const roms = {
  'OUTLAW': {
    name: 'Outlaw',
    rom: 'outlaw.ch8',
    description: 'Move with ASWD and fire with E.',
    cycles: "30",
  },
  'SLIPPERY_SLOPE': {
    name: 'Slippery Slope',
    rom: 'slipperyslope.ch8',
    description: 'Move with ASWD.',
    cycles: "200",
  },
  '15PUZZLE': {
    name: '15PUZZLE',
    rom: '15PUZZLE',
  },
  'BLINKY': {
    name: 'BLINKY',
    rom: 'BLINKY',
  },
  'BLITZ': {
    name: 'BLITZ',
    rom: 'BLITZ',
  },
  'BRIX': {
    name: 'BRIX',
    rom: 'BRIX',
  },
  'CONNECT4': {
    name: 'CONNECT4',
    rom: 'CONNECT4',
  },
  'GUESS': {
    name: 'GUESS',
    rom: 'GUESS',
  },
  'HIDDEN': {
    name: 'HIDDEN',
    rom: 'HIDDEN',
  },
  'INVADERS': {
    name: 'INVADERS',
    rom: 'INVADERS',
  },
  'KALEID': {
    name: 'KALEID',
    rom: 'KALEID',
  },
  'MAZE': {
    name: 'MAZE',
    rom: 'MAZE',
  },
  'MERLIN': {
    name: 'MERLIN',
    rom: 'MERLIN',
  },
  'MISSILE': {
    name: 'MISSILE',
    rom: 'MISSILE',
  },
  'PONG': {
    name: 'PONG',
    rom: 'PONG',
  },
  'PONG2': {
    name: 'PONG2',
    rom: 'PONG2',
  },
  'PUZZLE': {
    name: 'PUZZLE',
    rom: 'PUZZLE',
  },
  'SYZYGY': {
    name: 'SYZYGY',
    rom: 'SYZYGY',
  },
  'TANK': {
    name: 'TANK',
    rom: 'TANK',
  },
  'TETRIS': {
    name: 'TETRIS',
    rom: 'TETRIS',
  },
  'TICTAC': {
    name: 'TICTAC',
    rom: 'TICTAC',
  },
  'UFO': {
    name: 'UFO',
    rom: 'UFO',
  },
  'VBRIX': {
    name: 'VBRIX',
    rom: 'VBRIX',
  },
  'VERS': {
    name: 'VERS',
    rom: 'VERS',
  },
  'WIPEOFF': {
    name: 'WIPEOFF',
    rom: 'WIPEOFF',
  },
};

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

function loadRomFromSelect() {
  const select = document.getElementById('roms');
  const romKey = select.options[select.selectedIndex].value;
  const rom = roms[romKey];
  fetch(`roms/${rom.rom}`)
    .then(response => response.arrayBuffer())
    .then(buffer => {
      const uint8Array = new Uint8Array(buffer);
      document.getElementById('cycles').value = rom.cycles;
      document.getElementById('rom-description').innerText = rom.description;
      changeCycles();
      startRom(uint8Array);
    });
}

function loadRomFromText(text) {
  const uint8Array = new Uint8Array(
    text.replaceAll(',', ' ')
      .split(' ')
      .map(x => parseInt(x, 16))
  );
  startRom(uint8Array);
}

function loadRomFromUserText() {
  const text = document.getElementById('rom-text').value;
  loadRomFromText(text);
}

function startRom(rom) {
  emulator.loadRom(rom);
}

function beepListener() {
  const buffer = 0.5;
  if (this.currentTime > this.duration - buffer) {
    this.currentTime = 0;
    this.play();
  }
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

function changeCycles() {
  const cycles = parseInt(document.getElementById('cycles').value);
  if (isNaN(cycles)) {
    return;
  }
  emulator.setIpf(cycles);
}

function initRomSelect() {
  const select = document.getElementById('roms');
  select.innerHTML = '';
  for (const key in roms) {
    const option = document.createElement('option');
    option.value = key;
    option.text = roms[key].name;
    select.appendChild(option);
  }
}

function play_pause() {
  if (emulator.isPaused()) {
    emulator.resume();
  } else {
    emulator.pause();
  }
}

function getColor() {
  return document.getElementById('color').value.replace('#', '');
}

function setOnColorClick() {
  const color = getColor();
  if (!color.match(/^[0-9a-fA-F]{6}$/)) {
    return;
  }
  emulator.setOnColor(parseInt(color, 16));
}

function setOffColorClick() {
  const color = getColor();
  if (!color.match(/^[0-9a-fA-F]{6}$/)) {
    return;
  }
  emulator.setOffColor(parseInt(color, 16));
}

(async () => {
  // const canvas = document.getElementById('screen');
  // const container = document.getElementById('container');
  // initRomSelect();
  await loadWasm("main.wasm", "chip8-ui");
  // attachVisibilityListener();
  // const keyListeners = attachKeyListeners();
  // const romUploadListeners = attachRomUploadListeners();
  // const { resize } = attachResizeListener(canvas, container);
  // resize();
  // changeCycles();
  
  console.log('ChipStation Initialized');
})();
