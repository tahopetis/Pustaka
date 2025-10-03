const axios = require('axios');

const BASE_URL = 'http://localhost:8080/api/v1';

// Test script to verify the graph API fixes
async function testGraphFixes() {
    console.log('Testing Graph API Fixes...\n');

    try {
        // Test 1: CI Types filtering
        console.log('1. Testing CI Types filtering (should only return Server type nodes):');
        const serverTypesResponse = await axios.get(`${BASE_URL}/graph?ci_types=Server&limit=10`);
        console.log('   Status:', serverTypesResponse.status);
        console.log('   Nodes:', serverTypesResponse.data.nodes?.length || 0);
        console.log('   Node types:', [...new Set(serverTypesResponse.data.nodes?.map(n => n.type) || [])]);

        // Verify all nodes are of type Server
        const allServers = serverTypesResponse.data.nodes?.every(n => n.type === 'Server');
        console.log('   All nodes are Server type:', allServers);
        console.log('');

        // Test 2: Search functionality
        console.log('2. Testing search functionality:');
        const searchResponse = await axios.get(`${BASE_URL}/graph?search=test&limit=10`);
        console.log('   Status:', searchResponse.status);
        console.log('   Nodes found:', searchResponse.data.nodes?.length || 0);
        console.log('   Search term: "test"');
        console.log('');

        // Test 3: Empty graph (no filters)
        console.log('3. Testing empty graph (no filters):');
        const emptyResponse = await axios.get(`${BASE_URL}/graph?limit=5`);
        console.log('   Status:', emptyResponse.status);
        console.log('   Nodes:', emptyResponse.data.nodes?.length || 0);
        console.log('   Edges:', emptyResponse.data.edges?.length || 0);
        console.log('');

        // Test 4: Node expansion (network)
        console.log('4. Testing node expansion:');
        // First get a CI ID to test with
        const listResponse = await axios.get(`${BASE_URL}/ci?limit=1`);
        if (listResponse.data.cis && listResponse.data.cis.length > 0) {
            const testCI = listResponse.data.cis[0];
            console.log(`   Testing with CI: ${testCI.name} (${testCI.id})`);

            try {
                const networkResponse = await axios.get(`${BASE_URL}/ci/${testCI.id}/network?depth=2`);
                console.log('   Network Status:', networkResponse.status);
                console.log('   Network Nodes:', networkResponse.data.nodes?.length || 0);
                console.log('   Network Edges:', networkResponse.data.edges?.length || 0);

                if (networkResponse.data.nodes) {
                    const centerNode = networkResponse.data.nodes.find(n => n.id === testCI.id);
                    console.log('   Center node found:', !!centerNode);
                }
            } catch (networkError) {
                console.log('   Network Error:', networkError.response?.status, networkError.response?.data?.message || networkError.message);
            }
        } else {
            console.log('   No CIs found to test network expansion');
        }
        console.log('');

        // Test 5: Edge deduplication
        console.log('5. Testing edge deduplication:');
        const dedupResponse = await axios.get(`${BASE_URL}/graph?limit=20`);
        if (dedupResponse.data.edges) {
            const edgeMap = new Map();
            let duplicateCount = 0;

            dedupResponse.data.edges.forEach(edge => {
                const key = `${edge.source}-${edge.target}`;
                if (edgeMap.has(key)) {
                    duplicateCount++;
                } else {
                    edgeMap.set(key, true);
                }
            });

            console.log('   Total edges:', dedupResponse.data.edges.length);
            console.log('   Unique edges:', edgeMap.size);
            console.log('   Duplicates found:', duplicateCount);
            console.log('   Deduplication working:', duplicateCount === 0);
        }
        console.log('');

        // Test 6: Debug endpoint
        console.log('6. Testing debug endpoint:');
        try {
            const debugResponse = await axios.get(`${BASE_URL}/graph/debug`);
            console.log('   Debug Status:', debugResponse.status);
            console.log('   Debug results available:', !!debugResponse.data.test_results);

            if (debugResponse.data.test_results) {
                Object.entries(debugResponse.data.test_results).forEach(([test, result]) => {
                    console.log(`   ${test}: ${result.success ? 'PASS' : 'FAIL'}`);
                    if (result.error) {
                        console.log(`     Error: ${result.error}`);
                    }
                });
            }
        } catch (debugError) {
            console.log('   Debug Error:', debugError.response?.status, debugError.response?.data?.message || debugError.message);
        }

        console.log('\n✅ Graph API fixes testing completed!');

    } catch (error) {
        console.error('❌ Test failed:', error.response?.status, error.response?.data?.message || error.message);

        if (error.code === 'ECONNREFUSED') {
            console.log('\n💡 Make sure the application is running on port 8080');
            console.log('   Run: docker-compose up -d');
        }
    }
}

// Run the tests
testGraphFixes();