const { PrismaClient } = require('@prisma/client');

// Test multiple connection strings
const urls = [
    'postgresql://user:password@localhost:5433/ibos_db?schema=public',
    'postgresql://user:password@127.0.0.1:5433/ibos_db?schema=public',
    'postgresql://user@localhost:5433/ibos_db?schema=public',
    'postgresql://postgres@localhost:5433/ibos_db?schema=public',
];

async function test() {
    for (const url of urls) {
        const client = new PrismaClient({
            datasources: { db: { url } },
            log: [],
        });
        try {
            console.log('Testing:', url.substring(0, 50) + '...');
            await client.$connect();
            console.log('  -> SUCCESS!');
            await client.$disconnect();
            return;
        } catch (e) {
            console.log('  -> FAIL:', e.message.substring(0, 80));
            await client.$disconnect().catch(() => {});
        }
    }
}

test();
