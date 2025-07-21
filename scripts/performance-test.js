// Performance test script for 100k tasks

const axios = require('axios');
const API_URL = 'http://localhost:8080/api/v1';

async function generateTestData() {
  console.log('Generating 100k tasks...');
  
  const batchSize = 1000;
  const totalTasks = 100000;
  const zones = [5, 6, 7, 8, 9, 10]; // Zone IDs
  const trades = [1, 2, 3, 4, 5];
  
  let startDate = new Date('2025-08-01');
  
  for (let i = 0; i < totalTasks; i += batchSize) {
    const tasks = [];
    
    for (let j = 0; j < batchSize && i + j < totalTasks; j++) {
      tasks.push({
        project_id: 1,
        zone_id: zones[Math.floor(Math.random() * zones.length)],
        name: `Task ${i + j}`,
        start_date: startDate.toISOString().split('T')[0],
        duration: Math.floor(Math.random() * 5) + 1,
        trade_id: trades[Math.floor(Math.random() * trades.length)],
        status: 'planned',
        sequence_number: j
      });
      
      // Move date forward
      if (j % 10 === 0) {
        startDate.setDate(startDate.getDate() + 1);
      }
    }
    
    // Bulk insert
    try {
      await axios.post(`${API_URL}/tasks/bulk`, tasks);
      console.log(`Inserted ${i + tasks.length} tasks`);
    } catch (error) {
      console.error('Error inserting batch:', error.message);
    }
  }
}

async function testPerformance() {
  console.log('Testing API performance...');
  
  // Test 1: Get all tasks for a zone
  console.time('Get tasks for zone');
  const response1 = await axios.get(`${API_URL}/tasks?project_id=1&zone_ids=5`);
  console.timeEnd('Get tasks for zone');
  console.log(`Retrieved ${response1.data.length} tasks`);
  console.log(`Response time: ${response1.headers['x-response-time']}`);
  
  // Test 2: Get tasks with date range
  console.time('Get tasks with date range');
  const response2 = await axios.get(`${API_URL}/tasks?project_id=1&date_from=2025-08-01&date_to=2025-08-31`);
  console.timeEnd('Get tasks with date range');
  console.log(`Retrieved ${response2.data.length} tasks`);
  
  // Test 3: Bulk update
  console.time('Bulk update 100 tasks');
  const updates = response1.data.slice(0, 100).map(task => ({
    id: task.id,
    status: 'in-progress'
  }));
  await axios.patch(`${API_URL}/tasks/bulk`, updates);
  console.timeEnd('Bulk update 100 tasks');
}

// Run tests
(async () => {
  try {
    const args = process.argv.slice(2);
    
    if (args.includes('--generate')) {
      await generateTestData();
    }
    
    if (args.includes('--test') || args.length === 0) {
      await testPerformance();
    }
  } catch (error) {
    console.error('Test failed:', error);
  }
})();
