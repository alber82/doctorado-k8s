const express = require('express');
const fs = require('fs');
const fsp = fs.promises;
const path = require('path');
const os = require('os');
const { exec } = require('child_process');
const { v4: uuidv4 } = require('uuid');

const app = express();
const PORT = process.env.PORT || 3000;
const DATA_DIR = '/data';

// Asegurar que la carpeta de trabajo existe
if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR, { recursive: true });
}

// Quién atiende (útil para balanceo/LB)
app.get('/whoami', (_req, res) => res.json({ pod: os.hostname() }));

// Uso real del filesystem que contiene /data
app.get('/disk-usage', (_req, res) => {
    exec('df -k /data', (err, stdout) => {
        if (err) return res.status(500).send('Error al obtener uso de disco');
        const lines = stdout.trim().split('\n');
        if (lines.length < 2) return res.status(500).send('Salida df inesperada');
        const parts = lines[1].trim().split(/\s+/);
        const sizeKB = parseInt(parts[1], 10);
        const usedKB = parseInt(parts[2], 10);
        const availKB = parseInt(parts[3], 10);
        const usage = (usedKB / sizeKB) * 100;
        res.json({
            usage: Number(usage.toFixed(2)),
            sizeBytes: sizeKB * 1024,
            usedBytes: usedKB * 1024,
            availBytes: availKB * 1024
        });
    });
});

// Escribir en disco (archivos únicos)
app.post('/write', async (_req, res) => {
    try {
        const filename = `file_${uuidv4()}.dat`;
        const filepath = path.join(DATA_DIR, filename);
        const data = 'A'.repeat(1 * 1024 * 1024); // 1MB
        await fsp.writeFile(filepath, data);
        res.send(`Archivo ${filename} creado`);
    } catch (err) {
        res.status(500).send('Error al escribir en disco');
    }
});

// Leer un archivo aleatorio (robusto frente a ENOENT)
app.get('/read', async (_req, res) => {
    try {
        const dirents = await fsp.readdir(DATA_DIR, { withFileTypes: true });
        const onlyFiles = dirents.filter(d => d.isFile()).map(d => d.name);
        if (onlyFiles.length === 0) return res.status(500).send('No hay archivos para leer');

        for (let i = 0; i < Math.min(5, onlyFiles.length); i++) {
            const randomFile = onlyFiles[Math.floor(Math.random() * onlyFiles.length)];
            const filepath = path.join(DATA_DIR, randomFile);
            try {
                const data = await fsp.readFile(filepath);
                return res.send(`Archivo leído: ${randomFile}, tamaño: ${data.length} bytes`);
            } catch (e) {
                if (e.code !== 'ENOENT') throw e; // si no es "no existe", es un error real
            }
        }
        return res.status(500).send('No se pudo leer: archivos desaparecieron durante la lectura');
    } catch (_err) {
        return res.status(500).send('Error al leer el directorio');
    }
});

// Lock muy simple para serializar operaciones destructivas
let dirBusy = false;
const waiters = [];
async function withDirLock(fn) {
    if (dirBusy) await new Promise(r => waiters.push(r));
    dirBusy = true;
    try { return await fn(); }
    finally {
        dirBusy = false;
        const next = waiters.shift();
        if (next) next();
    }
}

// Borrar archivos más antiguos (robusto + lock)
app.delete('/delete', async (_req, res) => {
    try {
        await withDirLock(async () => {
            const dirents = await fsp.readdir(DATA_DIR, { withFileTypes: true });
            const infos = [];
            for (const d of dirents) {
                if (!d.isFile()) continue;
                const p = path.join(DATA_DIR, d.name);
                try {
                    const st = await fsp.stat(p);
                    infos.push({ name: d.name, mtime: st.mtimeMs });
                } catch (e) {
                    if (e.code !== 'ENOENT') throw e; // si ya no existe, lo ignoramos
                }
            }

            if (infos.length === 0) {
                res.send('No hay archivos para eliminar');
                return;
            }

            // Ordenar por fecha (más antiguo primero)
            infos.sort((a, b) => a.mtime - b.mtime);

            // Intentar borrar el primero disponible; si desapareció, probar siguientes
            let deleted = null;
            for (const info of infos) {
                const p = path.join(DATA_DIR, info.name);
                try {
                    await fsp.unlink(p);
                    deleted = info.name;
                    break;
                } catch (e) {
                    if (e.code !== 'ENOENT') throw e;
                }
            }

            if (deleted) res.send(`Archivo eliminado: ${deleted}`);
            else res.send('No se pudo eliminar ningún archivo (todos desaparecieron)');
        });
    } catch (_err) {
        res.status(500).send('Error al borrar archivo');
    }
});

// Iniciar el servidor
app.listen(PORT, () => console.log(`Servidor corriendo en el puerto ${PORT}`));