import { useState, useRef, useEffect } from 'react';
import './SearchableSelect.css';
import { LuSearch, LuX, LuChevronDown } from 'react-icons/lu';

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

interface SearchableSelectProps {
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  emptyPlaceholder?: string;
  label?: string;
  id?: string;
}

export const SearchableSelect: React.FC<SearchableSelectProps> = ({
  options,
  value,
  onChange,
  placeholder = 'Select...',
  emptyPlaceholder = 'No parent (root level)',
  label,
  id,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);

  const selectedOption = options.find(opt => opt.value === value);
  const displayValue = value === '' ? emptyPlaceholder : (selectedOption?.label || '');

  // Filter options based on search query
  const filteredOptions = options.filter(option =>
    option.label.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
        setSearchQuery('');
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      // Focus search input when dropdown opens
      setTimeout(() => searchInputRef.current?.focus(), 50);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  const handleSelect = (optionValue: string) => {
    onChange(optionValue);
    setIsOpen(false);
    setSearchQuery('');
  };

  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation();
    onChange('');
    setSearchQuery('');
  };

  return (
    <div className="searchable-select-wrapper">
      {label && <label htmlFor={id} className="searchable-select-label">{label}</label>}
      <div className="searchable-select" ref={containerRef}>
        <button
          type="button"
          className={`searchable-select-trigger ${isOpen ? 'open' : ''}`}
          onClick={() => setIsOpen(!isOpen)}
          id={id}
        >
          <span className="searchable-select-value">
            {displayValue || placeholder}
          </span>
          <div className="searchable-select-icons">
            {value && value !== '' && (
              <button
                type="button"
                className="searchable-select-clear"
                onClick={handleClear}
                aria-label="Clear selection"
              >
                <LuX size={16} />
              </button>
            )}
            <LuChevronDown className={`searchable-select-chevron ${isOpen ? 'rotate' : ''}`} size={16} />
          </div>
        </button>

        {isOpen && (
          <div className="searchable-select-dropdown">
            <div className="searchable-select-search">
              <LuSearch className="search-icon" size={16} />
              <input
                ref={searchInputRef}
                type="text"
                className="search-input"
                placeholder="Search..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onClick={(e) => e.stopPropagation()}
              />
            </div>

            <div className="searchable-select-options">
              {/* Empty/None option */}
              <button
                type="button"
                className={`searchable-select-option ${value === '' ? 'selected' : ''}`}
                onClick={() => handleSelect('')}
              >
                <span className="option-label">{emptyPlaceholder}</span>
              </button>

              {/* Filtered options */}
              {filteredOptions.length > 0 ? (
                filteredOptions.map((option) => (
                  <button
                    key={option.value}
                    type="button"
                    className={`searchable-select-option ${value === option.value ? 'selected' : ''} ${option.disabled ? 'disabled' : ''}`}
                    onClick={() => !option.disabled && handleSelect(option.value)}
                    disabled={option.disabled}
                  >
                    <span className="option-label">{option.label}</span>
                  </button>
                ))
              ) : (
                <div className="searchable-select-empty">
                  No results found for "{searchQuery}"
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default SearchableSelect;
