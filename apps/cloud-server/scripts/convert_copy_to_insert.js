const fs = require('fs');
const readline = require('readline');

async function processFile() {
    const fileStream = fs.createReadStream('C:\\Users\\tech\\Nembus\\apps\\cloud-server\\scripts\\extracted-data-copy-v2.sql');
    const outStream = fs.createWriteStream('C:\\Users\\tech\\Nembus\\apps\\cloud-server\\scripts\\init_all_data_dump_v2.sql');

    const rl = readline.createInterface({
        input: fileStream,
        crlfDelay: Infinity
    });

    let currentTable = null;
    let currentColumns = null;
    let inCopyBlock = false;

    for await (const line of rl) {
        if (line.startsWith('COPY ')) {
            const match = line.match(/COPY (.+?) \((.+?)\) FROM stdin;/);
            if (match) {
                currentTable = match[1];
                currentColumns = match[2];
                inCopyBlock = true;
                outStream.write(`\n-- =====================================================\n`);
                outStream.write(`-- Data for ${currentTable}\n`);
                outStream.write(`-- =====================================================\n`);
                continue;
            }
        }

        if (inCopyBlock) {
            if (line === '\\.') {
                inCopyBlock = false;
                currentTable = null;
                currentColumns = null;
                outStream.write('\n');
                continue;
            }
            
            const values = line.split('\t').map(v => {
                if (v === '\\N') return 'NULL';
                // Escape single quotes for SQL
                return "'" + v.replace(/'/g, "''") + "'";
            });
            
            outStream.write(`INSERT INTO ${currentTable} (${currentColumns}) VALUES (${values.join(', ')});\n`);
        } else {
            if (!line.startsWith('SET ') && !line.startsWith('SELECT pg_catalog') && line.trim() !== '') {
                outStream.write(line + '\n');
            }
        }
    }
    
    console.log('Conversion complete!');
}

processFile();
