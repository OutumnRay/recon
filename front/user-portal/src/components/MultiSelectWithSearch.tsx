import React, { useState, useMemo } from 'react';
import './MultiSelectWithSearch.css';

interface Option {
  id: string;
  name: string;
  description?: string;
  email?: string;
}

interface MultiSelectWithSearchProps {
  label: string;
  options: Option[];
  selectedIds: string[];
  onChange: (selectedIds: string[]) => void;
  placeholder?: string;
  emptyMessage?: string;
  disabled?: boolean;
  disabledIds?: string[];
}

export const MultiSelectWithSearch: React.FC<MultiSelectWithSearchProps> = ({
  label,
  options,
  selectedIds,
  onChange,
  placeholder = 'Поиск...',
  emptyMessage = 'Нет доступных элементов',
  disabled = false,
  disabledIds = [],
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);

  // Filter options based on search query
  const filteredOptions = useMemo(() => {
    if (!searchQuery.trim()) {
      return options;
    }

    const query = searchQuery.toLowerCase();
    return options.filter(
      (option) =>
        option.name.toLowerCase().includes(query) ||
        option.description?.toLowerCase().includes(query) ||
        option.email?.toLowerCase().includes(query)
    );
  }, [options, searchQuery]);

  // Get selected options for display
  const selectedOptions = useMemo(() => {
    return options.filter((opt) => selectedIds.includes(opt.id));
  }, [options, selectedIds]);

  const handleToggle = (id: string) => {
    if (disabledIds.includes(id)) return;

    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter((selectedId) => selectedId !== id));
    } else {
      onChange([...selectedIds, id]);
    }
  };

  const handleRemove = (id: string) => {
    if (disabledIds.includes(id)) return;
    onChange(selectedIds.filter((selectedId) => selectedId !== id));
  };

  const handleClearAll = () => {
    // Only remove items that are not disabled
    onChange(selectedIds.filter((id) => disabledIds.includes(id)));
    setSearchQuery('');
  };

  return (
    <div className="multi-select-container">
      <label className="multi-select-label">{label}</label>

      {/* Search and dropdown */}
      <div className="multi-select-search-wrapper">
        <input
          type="text"
          className="multi-select-search"
          placeholder={placeholder}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onFocus={() => setIsDropdownOpen(true)}
          disabled={disabled}
        />

        {selectedIds.length > 0 && !disabled && (
          <button
            type="button"
            className="clear-all-btn"
            onClick={handleClearAll}
            title="Очистить все"
          >
            ✕
          </button>
        )}

        {isDropdownOpen && (
          <>
            <div
              className="dropdown-overlay"
              onClick={() => setIsDropdownOpen(false)}
            />
            <div className="multi-select-dropdown">
              {filteredOptions.length === 0 ? (
                <div className="dropdown-empty">
                  {options.length === 0 ? emptyMessage : 'Ничего не найдено'}
                </div>
              ) : (
                <div className="dropdown-list">
                  {filteredOptions.map((option) => {
                    const isSelected = selectedIds.includes(option.id);
                    const isDisabled = disabledIds.includes(option.id);

                    return (
                      <label
                        key={option.id}
                        className={`dropdown-item ${isSelected ? 'selected' : ''} ${
                          isDisabled ? 'disabled' : ''
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleToggle(option.id)}
                          disabled={isDisabled}
                        />
                        <div className="dropdown-item-content">
                          <div className="dropdown-item-name">{option.name}</div>
                          {(option.email || option.description) && (
                            <div className="dropdown-item-meta">
                              {option.email && <span>{option.email}</span>}
                              {option.description && <span>{option.description}</span>}
                            </div>
                          )}
                        </div>
                      </label>
                    );
                  })}
                </div>
              )}
            </div>
          </>
        )}
      </div>

      {/* Selected items display */}
      {selectedOptions.length > 0 && (
        <div className="selected-items">
          <div className="selected-items-header">
            <span>Выбрано: {selectedOptions.length}</span>
          </div>
          <div className="selected-items-list">
            {selectedOptions.map((option) => {
              const isDisabled = disabledIds.includes(option.id);

              return (
                <div
                  key={option.id}
                  className={`selected-item ${isDisabled ? 'disabled' : ''}`}
                >
                  <div className="selected-item-content">
                    <span className="selected-item-name">{option.name}</span>
                    {(option.email || option.description) && (
                      <span className="selected-item-meta">
                        {option.email || option.description}
                      </span>
                    )}
                  </div>
                  {!isDisabled && !disabled && (
                    <button
                      type="button"
                      className="remove-item-btn"
                      onClick={() => handleRemove(option.id)}
                      title="Удалить"
                    >
                      ✕
                    </button>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
