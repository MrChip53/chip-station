const keyMap = {
  '1': 0x1,
  '2': 0x2,
  '3': 0x3,
  '4': 0xC,
  'q': 0x4,
  'w': 0x5,
  'e': 0x6,
  'r': 0xD,
  'a': 0x7,
  's': 0x8,
  'd': 0x9,
  'f': 0xE,
  'z': 0xA,
  'x': 0x0,
  'c': 0xB,
  'v': 0xF,
};

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

async function loadWasm() {
  if (WebAssembly && !WebAssembly.instantiateStreaming) { // polyfill
    WebAssembly.instantiateStreaming = async (resp, importObject) => {
      const source = await (await resp).arrayBuffer();
      return await WebAssembly.instantiate(source, importObject);
     };
  }  

  const go = new Go();

  const result = await WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject);
  go.run(result.instance);
}


function attachKeyListeners() {
  const keyDownListener = (event) => {
    const key = event.key;
    const keyName = keyMap[key];
    if (keyName) {
      setKeyState(keyName, 1);
    }
  };

  const keyUpListener = (event) => {
    const key = event.key;
    const keyName = keyMap[key];
    if (keyName) {
      setKeyState(keyName, 0);
    }
  };

  document.addEventListener('keydown', keyDownListener);
  document.addEventListener('keyup', keyUpListener);

  return { keyDownListener, keyUpListener };
}

function attachRomUploadListeners() {
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
    const file = event.dataTransfer.files[0];
    const reader = new FileReader();
    reader.onload = function(e) {
      const arrayBuffer = e.target.result;
      const uint8Array = new Uint8Array(arrayBuffer);
      startRom(uint8Array);
      document.body.style.backgroundColor = '#000';
    };
    reader.readAsArrayBuffer(file);
  };

  document.addEventListener('dragover', dragOverListener);
  document.addEventListener('dragleave', dragLeaveListener);
  document.addEventListener('drop', dropListener);

  return { dragOverListener, dragLeaveListener, dropListener };
}

function attachResizeListener(canvas, container) {
  const resize = () => {
    canvas.width = document.body.clientWidth*0.9;
    canvas.height = (document.body.clientHeight-container.clientHeight)*0.9;
  };

  window.addEventListener('resize', resize);

  return { resize };
}

function attachVisibilityListener() {
  let runningOnHide = false;

  const visibilityChangeListener = () => {
    if (document.hidden) {
      runningOnHide = !isPaused();
      if (runningOnHide) {
        console.log("Pausing on hide");
        pause();
      }
    } else {
      if (runningOnHide) {
        console.log("Resuming on show");
        resume();
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
  loadRom(rom);
}

function downloadRom() {
  const rom = getRom();
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
  setIpf(cycles);
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
  if (isPaused()) {
    resume();
  } else {
    pause();
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
  setOnColor(parseInt(color, 16));
}

function setOffColorClick() {
  const color = getColor();
  if (!color.match(/^[0-9a-fA-F]{6}$/)) {
    return;
  }
  setOffColor(parseInt(color, 16));
}

(async () => {
  const canvas = document.getElementById('screen');
  const container = document.getElementById('container');
  initRomSelect();
  await loadWasm();
  attachVisibilityListener();
  const keyListeners = attachKeyListeners();
  const romUploadListeners = attachRomUploadListeners();
  const { resize } = attachResizeListener(canvas, container);
  resize();
  changeCycles();
  
  console.log('ChipStation Initialized');
})();
