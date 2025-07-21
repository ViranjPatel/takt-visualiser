import React, { useState } from 'react';
import { Task } from '../types';
import axios from 'axios';

interface TaskEditorProps {
  task?: Task;
  onSave: () => void;
  onCancel: () => void;
}

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

const TaskEditor: React.FC<TaskEditorProps> = ({ task, onSave, onCancel }) => {
  const [formData, setFormData] = useState({
    name: task?.name || '',
    start_date: task?.start_date || new Date().toISOString().split('T')[0],
    duration: task?.duration || 1,
    status: task?.status || 'planned'
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      if (task) {
        // Update existing
        await axios.patch(`${API_URL}/tasks/${task.id}`, formData);
      } else {
        // Create new
        await axios.post(`${API_URL}/tasks`, {
          ...formData,
          project_id: 1,
          zone_id: 1 // TODO: Get from context
        });
      }
      onSave();
    } catch (error) {
      console.error('Failed to save task:', error);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ padding: '20px' }}>
      <div style={{ marginBottom: '10px' }}>
        <label>Task Name:</label>
        <input
          type="text"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          required
          style={{ width: '100%', padding: '5px' }}
        />
      </div>
      
      <div style={{ marginBottom: '10px' }}>
        <label>Start Date:</label>
        <input
          type="date"
          value={formData.start_date}
          onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
          required
          style={{ width: '100%', padding: '5px' }}
        />
      </div>
      
      <div style={{ marginBottom: '10px' }}>
        <label>Duration (days):</label>
        <input
          type="number"
          min="1"
          value={formData.duration}
          onChange={(e) => setFormData({ ...formData, duration: parseInt(e.target.value) })}
          required
          style={{ width: '100%', padding: '5px' }}
        />
      </div>
      
      <div style={{ marginBottom: '10px' }}>
        <label>Status:</label>
        <select
          value={formData.status}
          onChange={(e) => setFormData({ ...formData, status: e.target.value })}
          style={{ width: '100%', padding: '5px' }}
        >
          <option value="planned">Planned</option>
          <option value="in-progress">In Progress</option>
          <option value="completed">Completed</option>
        </select>
      </div>
      
      <div style={{ display: 'flex', gap: '10px' }}>
        <button type="submit" style={{ flex: 1, padding: '8px' }}>Save</button>
        <button type="button" onClick={onCancel} style={{ flex: 1, padding: '8px' }}>Cancel</button>
      </div>
    </form>
  );
};

export default TaskEditor;
