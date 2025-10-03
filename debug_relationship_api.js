#!/usr/bin/env node

const axios = require('axios');

const API_BASE = 'http://localhost:8080/api/v1';

// Test authentication first
async function testAuth() {
  try {
    console.log('🔐 Testing authentication...');
    const response = await axios.post(`${API_BASE}/auth/login`, {
      username: 'admin',
      password: 'Admin@123'  // Correct password from docker-compose.yml
    });

    console.log('✅ Authentication successful');
    console.log('User:', response.data.user.username);
    console.log('Email:', response.data.user.email);
    console.log('Roles:', response.data.user.roles.map(r => r.name));
    console.log('Permissions:', response.data.user.permissions);

    return response.data.access_token;
  } catch (error) {
    console.error('❌ Authentication failed:', error.response?.data || error.message);
    if (error.response?.status === 401) {
      console.log('Check if the API server is running and accessible');
    }
    process.exit(1);
  }
}

// Test relationships API
async function testRelationships(token) {
  try {
    console.log('\n🔗 Testing relationships API...');

    const config = {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    };

    // Test list relationships
    console.log('Fetching relationships...');
    const response = await axios.get(`${API_BASE}/relationships`, config);

    console.log('✅ Relationships API call successful');
    console.log('Total relationships:', response.data.total);
    console.log('Relationships in response:', response.data.relationships.length);

    if (response.data.relationships.length > 0) {
      console.log('Sample relationship:', {
        id: response.data.relationships[0].id,
        source_id: response.data.relationships[0].source_id,
        target_id: response.data.relationships[0].target_id,
        type: response.data.relationships[0].relationship_type
      });
    }

    // Test CIs to see if there are any CIs to create relationships with
    console.log('\n🖥️ Testing CIs API...');
    const ciResponse = await axios.get(`${API_BASE}/ci?limit=10`, config);
    console.log('Total CIs:', ciResponse.data.total);

    if (ciResponse.data.cis.length > 0) {
      console.log('Sample CIs:', ciResponse.data.cis.map(ci => ({
        id: ci.id,
        name: ci.name,
        type: ci.ci_type
      })));

      // Test creating a relationship if we have at least 2 CIs
      if (ciResponse.data.cis.length >= 2) {
        console.log('\n🔗 Testing relationship creation...');
        const sourceCI = ciResponse.data.cis[0];
        const targetCI = ciResponse.data.cis[1];

        try {
          const createResponse = await axios.post(`${API_BASE}/relationships`, {
            source_id: sourceCI.id,
            target_id: targetCI.id,
            relationship_type: 'depends_on',
            attributes: { test: 'debug_relationship' }
          }, config);

          console.log('✅ Relationship created successfully');
          console.log('New relationship:', {
            id: createResponse.data.id,
            source: sourceCI.name,
            target: targetCI.name,
            type: createResponse.data.relationship_type
          });

          // Test listing relationships again
          console.log('\n🔗 Testing relationships after creation...');
          const listResponse = await axios.get(`${API_BASE}/relationships`, config);
          console.log('Total relationships after creation:', listResponse.data.total);

        } catch (createError) {
          if (createError.response?.status === 400 && createError.response?.data?.error?.includes('already exists')) {
            console.log('ℹ️  Relationship already exists between these CIs');
          } else {
            console.error('❌ Failed to create relationship:', createError.response?.data || createError.message);
          }
        }
      }
    } else {
      console.log('ℹ️  No CIs found - need to create CIs first to test relationships');
    }

  } catch (error) {
    console.error('❌ Relationships API test failed:', error.response?.data || error.message);

    if (error.response?.status === 403) {
      console.log('🔒 Permission denied - checking if user has relationship:read permission');
    } else if (error.response?.status === 401) {
      console.log('🔑 Token may have expired');
    }
  }
}

// Test user permissions
async function testUserPermissions(token) {
  try {
    console.log('\n👤 Testing user permissions...');

    const config = {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    };

    const response = await axios.get(`${API_BASE}/me`, config);

    console.log('✅ User permissions retrieved');
    console.log('Has relationship:read permission:', response.data.user.permissions.includes('relationship:read'));
    console.log('Has relationship:create permission:', response.data.user.permissions.includes('relationship:create'));

    return response.data.user.permissions;
  } catch (error) {
    console.error('❌ Failed to get user permissions:', error.response?.data || error.message);
    return [];
  }
}

// Main test function
async function runTests() {
  console.log('🧪 Starting API diagnostics...\n');

  try {
    const token = await testAuth();
    const permissions = await testUserPermissions(token);
    await testRelationships(token);

    console.log('\n✅ All tests completed');
    console.log('\n📝 Summary:');
    console.log('- API server is running and accessible');
    console.log('- Authentication is working');
    console.log(`- User has ${permissions.length} permissions`);
    console.log(`- User has relationship:read permission: ${permissions.includes('relationship:read')}`);
    console.log(`- User has relationship:create permission: ${permissions.includes('relationship:create')}`);

  } catch (error) {
    console.error('\n❌ Test suite failed:', error.message);
    process.exit(1);
  }
}

runTests();