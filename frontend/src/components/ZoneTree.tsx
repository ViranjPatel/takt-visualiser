import React from 'react';
import { Zone } from '../types';

interface ZoneTreeProps {
  zones: Zone[];
  selectedZones: number[];
  onSelectionChange: (zones: number[]) => void;
}

const ZoneTree: React.FC<ZoneTreeProps> = ({ zones, selectedZones, onSelectionChange }) => {
  const handleZoneClick = (zoneId: number, event: React.MouseEvent) => {
    event.stopPropagation();
    
    if (event.ctrlKey || event.metaKey) {
      // Multi-select
      if (selectedZones.includes(zoneId)) {
        onSelectionChange(selectedZones.filter(id => id !== zoneId));
      } else {
        onSelectionChange([...selectedZones, zoneId]);
      }
    } else {
      // Single select
      onSelectionChange([zoneId]);
    }
  };

  const renderZone = (zone: Zone) => {
    const isSelected = selectedZones.includes(zone.id);
    
    return (
      <div key={zone.id}>
        <div
          className={`zone-item ${isSelected ? 'selected' : ''}`}
          onClick={(e) => handleZoneClick(zone.id, e)}
          style={{
            paddingLeft: `${zone.level * 20}px`,
            backgroundColor: isSelected ? '#e3f2fd' : 'transparent',
            fontWeight: isSelected ? 'bold' : 'normal'
          }}
        >
          {zone.name}
        </div>
        {zone.children && zone.children.map(child => renderZone(child))}
      </div>
    );
  };

  return (
    <div className="zone-tree">
      {zones.map(zone => renderZone(zone))}
    </div>
  );
};

export default ZoneTree;
