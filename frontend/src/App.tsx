import React, { useState, useEffect, useCallback } from 'react';
import { DayPilotScheduler } from '@daypilot/daypilot-lite-react';
import axios from 'axios';
import ZoneTree from './components/ZoneTree';
import TaskEditor from './components/TaskEditor';
import { Task, Zone } from './types';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

function App() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [zones, setZones] = useState<Zone[]>([]);
  const [selectedZones, setSelectedZones] = useState<number[]>([]);
  const [loading, setLoading] = useState(true);
  const [ws, setWs] = useState<WebSocket | null>(null);

  // Load zones
  useEffect(() => {
    loadZones();
  }, []);

  // Load tasks when zones change
  useEffect(() => {
    if (selectedZones.length > 0) {
      loadTasks();
    }
  }, [selectedZones]);

  // WebSocket connection
  useEffect(() => {
    const websocket = new WebSocket('ws://localhost:8080/api/v1/ws');
    
    websocket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'task_update') {
        // Reload single task
        loadTask(data.task_id);
      }
    };

    setWs(websocket);

    return () => {
      websocket.close();
    };
  }, []);

  const loadZones = async () => {
    try {
      const response = await axios.get(`${API_URL}/zones/1/tree`);
      setZones(response.data);
    } catch (error) {
      console.error('Failed to load zones:', error);
    }
  };

  const loadTasks = async () => {
    setLoading(true);
    try {
      const params = {
        project_id: 1,
        zone_ids: selectedZones.join(',')
      };
      const response = await axios.get(`${API_URL}/tasks`, { params });
      setTasks(response.data);
    } catch (error) {
      console.error('Failed to load tasks:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadTask = async (taskId: number) => {
    try {
      const response = await axios.get(`${API_URL}/tasks/${taskId}`);
      setTasks(prev => {
        const index = prev.findIndex(t => t.id === taskId);
        if (index >= 0) {
          const updated = [...prev];
          updated[index] = response.data;
          return updated;
        }
        return prev;
      });
    } catch (error) {
      console.error('Failed to load task:', error);
    }
  };

  const handleTaskMove = useCallback(async (args: any) => {
    const taskId = args.e.data.id;
    const newStart = args.newStart;
    const newResource = args.newResource;

    try {
      await axios.patch(`${API_URL}/tasks/${taskId}`, {
        start_date: newStart.toString('yyyy-MM-dd'),
        zone_id: parseInt(newResource)
      });
    } catch (error) {
      console.error('Failed to update task:', error);
      // Reload to revert
      loadTasks();
    }
  }, []);

  const handleTaskResize = useCallback(async (args: any) => {
    const taskId = args.e.data.id;
    const newStart = args.newStart;
    const newEnd = args.newEnd;
    const duration = Math.ceil((newEnd - newStart) / (1000 * 60 * 60 * 24));

    try {
      await axios.patch(`${API_URL}/tasks/${taskId}`, {
        start_date: newStart.toString('yyyy-MM-dd'),
        duration: duration
      });
    } catch (error) {
      console.error('Failed to update task:', error);
      loadTasks();
    }
  }, []);

  const schedulerConfig = {
    viewType: 'Days',
    days: 90,
    cellWidth: 40,
    rowHeaderWidth: 200,
    eventHeight: 25,
    headerHeight: 25,
    onEventMoved: handleTaskMove,
    onEventResized: handleTaskResize,
    treeEnabled: true,
    treePreventParentUsage: true
  };

  return (
    <div className="app">
      <header className="header">
        <h1>Takt Visualiser</h1>
      </header>
      <div className="content">
        <aside className="sidebar">
          <h3>Zones</h3>
          <ZoneTree 
            zones={zones}
            selectedZones={selectedZones}
            onSelectionChange={setSelectedZones}
          />
        </aside>
        <main className="main">
          {loading ? (
            <div>Loading...</div>
          ) : (
            <div className="scheduler-container">
              <DayPilotScheduler
                {...schedulerConfig}
                events={tasks.map(task => ({
                  id: task.id,
                  text: task.name,
                  start: task.start_date,
                  end: new Date(new Date(task.start_date).getTime() + task.duration * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
                  resource: task.zone_id,
                  cssClass: `trade-color-${task.trade_id || 1} status-${task.status}`
                }))}
                resources={flattenZones(zones)}
              />
            </div>
          )}
        </main>
      </div>
    </div>
  );
}

// Helper function to flatten zone tree for scheduler
function flattenZones(zones: Zone[], level = 0): any[] {
  let result: any[] = [];
  for (const zone of zones) {
    result.push({
      id: zone.id,
      name: zone.name,
      expanded: true,
      marginLeft: level * 20
    });
    if (zone.children && zone.children.length > 0) {
      result = result.concat(flattenZones(zone.children, level + 1));
    }
  }
  return result;
}

export default App;
