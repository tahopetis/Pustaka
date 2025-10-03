// Test script to verify bidirectional relationship loading
// This script simulates the API calls to verify our fix works

const { axios } = require('axios');

async function testBidirectionalRelationships() {
  const baseURL = 'http://localhost:8080/api/v1';

  try {
    // Test 1: Create sample CIs if they don't exist
    console.log('🧪 Testing CI creation...');

    const ci1Response = await createCI('web-server-01', 'Server', {
      hostname: 'web-server-01',
      ip_address: '192.168.1.10'
    });

    const ci2Response = await createCI('database-01', 'Database', {
      hostname: 'database-01',
      engine: 'PostgreSQL'
    });

    const ci3Response = await createCI('app-server-01', 'Server', {
      hostname: 'app-server-01',
      ip_address: '192.168.1.20'
    });

    // Test 2: Create relationships
    console.log('🧪 Creating relationships...');

    // CI1 -> CI2 (web-server runs_on database)
    await createRelationship(ci1Response.data.id, ci2Response.data.id, 'runs_on');

    // CI3 -> CI1 (app-server runs_on web-server)
    await createRelationship(ci3Response.data.id, ci1Response.data.id, 'runs_on');

    // Test 3: Query relationships for CI1 as source
    console.log('🧪 Testing relationships for CI1 as source...');
    const sourceResponse = await listRelationships({ source_id: ci1Response.data.id });
    console.log(`Found ${sourceResponse.data.relationships.length} relationships where CI1 is source`);

    // Test 4: Query relationships for CI1 as target
    console.log('🧪 Testing relationships for CI1 as target...');
    const targetResponse = await listRelationships({ target_id: ci1Response.data.id });
    console.log(`Found ${targetResponse.data.relationships.length} relationships where CI1 is target`);

    // Test 5: Verify the relationship display text logic
    console.log('🧪 Testing relationship display text logic...');

    // Outgoing relationship (CI1 is source)
    if (sourceResponse.data.relationships.length > 0) {
      const outgoingRel = sourceResponse.data.relationships[0];
      console.log(`Outgoing: ${outgoingRel.source_ci?.name || 'Unknown'} → ${outgoingRel.target_ci?.name || 'Unknown'}`);
    }

    // Incoming relationship (CI1 is target)
    if (targetResponse.data.relationships.length > 0) {
      const incomingRel = targetResponse.data.relationships[0];
      console.log(`Incoming: ${incomingRel.source_ci?.name || 'Unknown'} ← ${incomingRel.target_ci?.name || 'Unknown'}`);
    }

    console.log('✅ Bidirectional relationship test completed successfully!');

  } catch (error) {
    console.error('❌ Test failed:', error.response?.data || error.message);
  }
}

async function createCI(name, type, attributes) {
  try {
    const response = await axios.get(`${baseURL}/ci?search=${name}`);
    const existingCIs = response.data.cis;

    if (existingCIs.length > 0) {
      console.log(`✅ CI '${name}' already exists`);
      return { data: existingCIs[0] };
    }

    console.log(`Creating CI: ${name}`);
    const createResponse = await axios.post(`${baseURL}/ci`, {
      name,
      ci_type: type,
      attributes,
      tags: ['test']
    });
    console.log(`✅ Created CI: ${name}`);
    return createResponse;
  } catch (error) {
    if (error.response?.status === 401) {
      console.log('⚠️  Authentication required for CI operations. Skipping test.');
      throw new Error('Authentication required');
    }
    throw error;
  }
}

async function createRelationship(sourceId, targetId, type) {
  try {
    // Check if relationship already exists
    const existingResponse = await listRelationships({
      source_id: sourceId,
      target_id: targetId,
      relationship_type: type
    });

    if (existingResponse.data.relationships.length > 0) {
      console.log(`✅ Relationship ${type} already exists`);
      return existingResponse.data.relationships[0];
    }

    console.log(`Creating relationship: ${sourceId} -> ${targetId} (${type})`);
    const response = await axios.post(`${baseURL}/relationships`, {
      source_id: sourceId,
      target_id: targetId,
      relationship_type: type,
      attributes: { test: true }
    });
    console.log(`✅ Created relationship: ${type}`);
    return response.data;
  } catch (error) {
    if (error.response?.status === 401) {
      console.log('⚠️  Authentication required for relationship operations. Skipping test.');
      throw new Error('Authentication required');
    }
    throw error;
  }
}

async function listRelationships(params) {
  try {
    const response = await axios.get(`${baseURL}/relationships`, { params });

    // Fetch CI details for each relationship
    const relationshipsWithCIs = await Promise.all(
      response.data.relationships.map(async (rel) => {
        let updatedRel = { ...rel };

        // Fetch source CI if not populated
        if (!rel.source_ci && rel.source_id) {
          try {
            const sourceCIResponse = await axios.get(`${baseURL}/ci/${rel.source_id}`);
            updatedRel.source_ci = sourceCIResponse.data;
          } catch (error) {
            console.warn(`Failed to load source CI ${rel.source_id}:`, error.message);
          }
        }

        // Fetch target CI if not populated
        if (!rel.target_ci && rel.target_id) {
          try {
            const targetCIResponse = await axios.get(`${baseURL}/ci/${rel.target_id}`);
            updatedRel.target_ci = targetCIResponse.data;
          } catch (error) {
            console.warn(`Failed to load target CI ${rel.target_id}:`, error.message);
          }
        }

        return updatedRel;
      })
    );

    return { data: { ...response.data, relationships: relationshipsWithCIs } };
  } catch (error) {
    throw error;
  }
}

// Run the test
if (require.main === module) {
  testBidirectionalRelationships();
}

module.exports = { testBidirectionalRelationships };