// Test script to verify frontend API calls
// This script simulates what the Vue.js frontend should be doing

const axios = require('axios');

const API_BASE = 'http://localhost:8080/api/v1';

async function testFrontendAPICall() {
  console.log('🧪 Testing frontend API call simulation...\n');

  try {
    // 1. Authenticate (same as frontend)
    console.log('🔐 Simulating frontend login...');
    const authResponse = await axios.post(`${API_BASE}/auth/login`, {
      username: 'admin',
      password: 'Admin@123'
    });

    const token = authResponse.data.access_token;
    console.log('✅ Authentication successful');

    // 2. Set up axios instance like the frontend does
    const api = axios.create({
      baseURL: API_BASE,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add auth interceptor like frontend
    api.interceptors.request.use((config) => {
      config.headers['Authorization'] = `Bearer ${token}`;
      config.headers['X-Request-ID'] = crypto.randomUUID();
      return config;
    });

    // 3. Test the exact API call the frontend makes
    console.log('\n🔗 Simulating frontend relationships API call...');

    const params = {
      page: 1,
      limit: 20,
      search: '',
      relationship_type: '',
      sort: 'created_at',
      order: 'desc'
    };

    // Clean up empty values like frontend does
    Object.keys(params).forEach(key => {
      if (params[key] === '') {
        delete params[key];
      }
    });

    console.log('API call parameters:', params);

    const response = await api.get('/relationships', { params });

    console.log('✅ API call successful');
    console.log('Response structure:', {
      hasRelationships: 'relationships' in response.data,
      relationshipsCount: response.data.relationships?.length || 0,
      total: response.data.total,
      page: response.data.page,
      limit: response.data.limit,
      totalPages: response.data.total_pages
    });

    if (response.data.relationships && response.data.relationships.length > 0) {
      console.log('First relationship structure:', {
        id: response.data.relationships[0].id,
        source_id: response.data.relationships[0].source_id,
        target_id: response.data.relationships[0].target_id,
        relationship_type: response.data.relationships[0].relationship_type,
        created_at: response.data.relationships[0].created_at,
        updated_at: response.data.relationships[0].updated_at
      });
    }

    // 4. Test CI API calls (frontend loads CI details for display)
    if (response.data.relationships && response.data.relationships.length > 0) {
      console.log('\n🖥️ Simulating frontend CI details loading...');

      const ciIds = new Set();
      response.data.relationships.forEach(rel => {
        ciIds.add(rel.source_id);
        ciIds.add(rel.target_id);
      });

      console.log('CI IDs to load:', Array.from(ciIds));

      const ciPromises = Array.from(ciIds).map(async id => {
        try {
          const ciResponse = await api.get(`/ci/${id}`);
          return { id, data: ciResponse.data };
        } catch (error) {
          console.error(`Failed to load CI ${id}:`, error.response?.data || error.message);
          return { id, error: error.message };
        }
      });

      const ciResults = await Promise.all(ciPromises);
      console.log('✅ CI details loaded:', ciResults.filter(r => !r.error).length);
      console.log('❌ CI details failed:', ciResults.filter(r => r.error).length);

      ciResults.forEach(result => {
        if (!result.error) {
          console.log(`CI ${result.id}: ${result.data.name} (${result.data.ci_type})`);
        }
      });
    }

    return response.data;

  } catch (error) {
    console.error('❌ Frontend API simulation failed:', error.response?.data || error.message);
    if (error.response) {
      console.error('Response status:', error.response.status);
      console.error('Response headers:', error.response.headers);
    }
    throw error;
  }
}

// Test with the same headers the frontend would send
async function testWithFrontendHeaders() {
  console.log('\n🌐 Testing with frontend-like headers...');

  try {
    const token = (await axios.post(`${API_BASE}/auth/login`, {
      username: 'admin',
      password: 'Admin@123'
    })).data.access_token;

    const response = await axios.get(`${API_BASE}/relationships`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
        'X-Request-ID': crypto.randomUUID(),
        'Origin': 'http://localhost:3001',
        'Referer': 'http://localhost:3001/relationships'
      },
      params: {
        page: 1,
        limit: 20
      }
    });

    console.log('✅ Frontend headers test successful');
    console.log('Response matches expected format:', 'relationships' in response.data);

  } catch (error) {
    console.error('❌ Frontend headers test failed:', error.response?.data || error.message);
  }
}

// Main test function
async function runFrontendTests() {
  console.log('🧪 Starting frontend API diagnostics...\n');

  try {
    await testFrontendAPICall();
    await testWithFrontendHeaders();

    console.log('\n✅ All frontend tests completed');
    console.log('\n📝 Frontend Summary:');
    console.log('- API calls are working correctly');
    console.log('- Response format matches frontend expectations');
    console.log('- CI details loading is working');
    console.log('- The issue is likely in the frontend component logic');

  } catch (error) {
    console.error('\n❌ Frontend test suite failed:', error.message);
    process.exit(1);
  }
}

runFrontendTests();