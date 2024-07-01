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

function loadTank() {
  loadRomFromText('0x12 0x34 0x10 0x54 0x7C 0x6C 0x7C 0x7C 0x44 0x7C 0x7C 0x6C 0x7C 0x54 0x10 0x00 0xFC 0x78 0x6E 0x78 0xFC 0x00 0x3F 0x1E 0x76 0x1E 0x3F 0x00 0x72 0xFF 0xA2 0x02 0x00 0xEE 0x72 0x01 0xA2 0x08 0x00 0xEE 0x71 0xFF 0xA2 0x15 0x00 0xEE 0x71 0x01 0xA2 0x0F 0x00 0xEE 0x61 0x20 0x62 0x10 0xA2 0x02 0xD1 0x27 0xF0 0x0A 0xD1 0x27 0x40 0x02 0x22 0x1C 0x40 0x04 0x22 0x28 0x40 0x06 0x22 0x2E 0x40 0x08 0x22 0x22 0x12 0x3A');
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

(async () => {
  const canvas = document.getElementById('screen');
  const container = document.getElementById('container');
  initRomSelect();
  await loadWasm();
  const keyListeners = attachKeyListeners();
  const romUploadListeners = attachRomUploadListeners();
  const { resize } = attachResizeListener(canvas, container);
  resize();
  changeCycles();
  
  console.log('ChipStation Initialized');
})();
