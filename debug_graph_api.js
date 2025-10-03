#!/usr/bin/env node

const axios = require('axios');

const API_BASE = 'http://localhost:8080/api/v1';

// Test authentication first
async function testAuth() {
  try {
    console.log('🔐 Testing authentication...');
    const response = await axios.post(`${API_BASE}/auth/login`, {
      username: 'admin',
      password: 'Admin@123'
    });

    console.log('✅ Authentication successful');
    return response.data.access_token;
  } catch (error) {
    console.error('❌ Authentication failed:', error.response?.data || error.message);
    process.exit(1);
  }
}

// Test graph API with different filters
async function testGraphAPI(token) {
  try {
    console.log('\n🔍 Testing graph API...');

    const config = {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    };

    // Test 1: Get all graph data (no filters)
    console.log('\n1️⃣ Testing graph data without filters...');
    const response1 = await axios.get(`${API_BASE}/graph`, config);
    console.log('✅ Graph data (no filters):');
    console.log('  Nodes:', response1.data.nodes.length);
    console.log('  Edges:', response1.data.edges.length);
    if (response1.data.nodes.length > 0) {
      console.log('  Sample node:', {
        id: response1.data.nodes[0].id,
        name: response1.data.nodes[0].name,
        type: response1.data.nodes[0].type
      });
    }
    if (response1.data.edges.length > 0) {
      console.log('  Sample edge:', {
        id: response1.data.edges[0].id,
        source: response1.data.edges[0].source,
        target: response1.data.edges[0].target,
        type: response1.data.edges[0].relationship_type
      });
    }

    // Test 2: CI types filtering
    console.log('\n2️⃣ Testing CI types filtering...');
    const response2 = await axios.get(`${API_BASE}/graph?ci_types=Server`, config);
    console.log('✅ Graph data (ci_types=Server):');
    console.log('  Nodes:', response2.data.nodes.length);
    console.log('  Edges:', response2.data.edges.length);

    const response3 = await axios.get(`${API_BASE}/graph?ci_types=Application`, config);
    console.log('✅ Graph data (ci_types=Application):');
    console.log('  Nodes:', response3.data.nodes.length);
    console.log('  Edges:', response3.data.edges.length);

    // Test 3: Search functionality
    console.log('\n3️⃣ Testing search functionality...');
    const response4 = await axios.get(`${API_BASE}/graph?search=test`, config);
    console.log('✅ Graph data (search=test):');
    console.log('  Nodes:', response4.data.nodes.length);
    console.log('  Edges:', response4.data.edges.length);

    // Test 4: Empty graph with search (should return no nodes)
    console.log('\n4️⃣ Testing empty graph search...');
    const response5 = await axios.get(`${API_BASE}/graph?search=nonexistent`, config);
    console.log('✅ Graph data (search=nonexistent):');
    console.log('  Nodes:', response5.data.nodes.length);
    console.log('  Edges:', response5.data.edges.length);

    // Test 5: Limit parameter
    console.log('\n5️⃣ Testing limit parameter...');
    const response6 = await axios.get(`${API_BASE}/graph?limit=1`, config);
    console.log('✅ Graph data (limit=1):');
    console.log('  Nodes:', response6.data.nodes.length);
    console.log('  Edges:', response6.data.edges.length);

    // Test 6: Multiple CI types
    console.log('\n6️⃣ Testing multiple CI types...');
    const response7 = await axios.get(`${API_BASE}/graph?ci_types=Server&ci_types=Application`, config);
    console.log('✅ Graph data (multiple CI types):');
    console.log('  Nodes:', response7.data.nodes.length);
    console.log('  Edges:', response7.data.edges.length);

    // Check for duplicate edges
    console.log('\n🔍 Checking for duplicate edges...');
    const allEdges = response1.data.edges;
    const edgeMap = new Map();
    let duplicateCount = 0;

    allEdges.forEach(edge => {
      const key = `${edge.source}-${edge.target}-${edge.relationship_type}`;
      if (edgeMap.has(key)) {
        duplicateCount++;
        console.log(`❌ Duplicate edge found: ${key}`);
      } else {
        edgeMap.set(key, edge);
      }
    });

    console.log(`✅ Duplicate analysis: ${duplicateCount} duplicates found out of ${allEdges.length} edges`);

    return {
      totalNodes: response1.data.nodes.length,
      totalEdges: response1.data.edges.length,
      duplicateCount: duplicateCount
    };

  } catch (error) {
    console.error('❌ Graph API test failed:', error.response?.data || error.message);
    if (error.response?.status === 500) {
      console.log('💡 Server error - check backend logs for details');
    }
    throw error;
  }
}

// Test node expansion
async function testNodeExpansion(token) {
  try {
    console.log('\n🔗 Testing node expansion...');

    const config = {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    };

    // First, get a list of CIs to test with
    const ciResponse = await axios.get(`${API_BASE}/ci?limit=5`, config);
    if (ciResponse.data.cis.length === 0) {
      console.log('ℹ️  No CIs found for node expansion test');
      return;
    }

    const testCI = ciResponse.data.cis[0];
    console.log(`Testing node expansion for CI: ${testCI.name} (${testCI.id})`);

    // Test CI network with different depths
    for (let depth = 1; depth <= 3; depth++) {
      console.log(`\n📊 Testing network depth ${depth}...`);
      try {
        const networkResponse = await axios.get(`${API_BASE}/ci/${testCI.id}/network?depth=${depth}`, config);
        console.log(`✅ Network depth ${depth}:`);
        console.log(`  Nodes: ${networkResponse.data.nodes.length}`);
        console.log(`  Edges: ${networkResponse.data.edges.length}`);
      } catch (error) {
        console.log(`❌ Network depth ${depth} failed:`, error.response?.data?.error || error.message);
      }
    }

  } catch (error) {
    console.error('❌ Node expansion test failed:', error.response?.data || error.message);
  }
}

// Main test function
async function runTests() {
  console.log('🧪 Starting Graph API diagnostics...\n');

  try {
    const token = await testAuth();
    const graphResults = await testGraphAPI(token);
    await testNodeExpansion(token);

    console.log('\n✅ All graph tests completed');
    console.log('\n📝 Summary:');
    console.log(`- Total nodes in graph: ${graphResults.totalNodes}`);
    console.log(`- Total edges in graph: ${graphResults.totalEdges}`);
    console.log(`- Duplicate edges found: ${graphResults.duplicateCount}`);

    if (graphResults.duplicateCount > 0) {
      console.log('⚠️  ISSUE FOUND: Duplicate edges are being returned');
    }

    console.log('\n🔧 Issues to investigate:');
    console.log('1. CI types filtering - verify it works correctly');
    console.log('2. Duplicate relationships - check Neo4j query logic');
    console.log('3. Empty graph with search - should return empty results');
    console.log('4. Node expansion with max nodes limit (100)');

  } catch (error) {
    console.error('\n❌ Graph API test suite failed:', error.message);
    process.exit(1);
  }
}

runTests();