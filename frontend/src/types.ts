export interface Task {
  id: number;
  project_id: number;
  zone_id: number;
  name: string;
  start_date: string;
  duration: number;
  trade_id?: number;
  status: string;
  sequence_number?: number;
  color?: string;
  updated_at: string;
}

export interface Zone {
  id: number;
  project_id: number;
  parent_id?: number;
  name: string;
  level: number;
  path: string;
  children?: Zone[];
}

export interface Trade {
  id: number;
  name: string;
  color: string;
  project_id: number;
}
