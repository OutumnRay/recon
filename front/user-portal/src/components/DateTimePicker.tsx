import { forwardRef } from 'react';
import { LuCalendar } from 'react-icons/lu';
import './DateTimePicker.css';

interface DateTimePickerProps {
  id?: string;
  value: string;
  onChange: (value: string) => void;
  className?: string;
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
  min?: string;
  max?: string;
  type?: 'datetime-local' | 'date' | 'time';
}

/**
 * Custom styled DateTime Picker component matching site design
 * Wraps native datetime-local/date/time inputs with enhanced styling
 */
export const DateTimePicker = forwardRef<HTMLInputElement, DateTimePickerProps>(
  (
    {
      id,
      value,
      onChange,
      className = '',
      placeholder,
      required = false,
      disabled = false,
      min,
      max,
      type = 'datetime-local',
    },
    ref
  ) => {
    return (
      <div className={`datetime-picker-wrapper ${className}`}>
        <LuCalendar className="datetime-picker-icon" />
        <input
          ref={ref}
          id={id}
          type={type}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="datetime-picker-input"
          placeholder={placeholder}
          required={required}
          disabled={disabled}
          min={min}
          max={max}
        />
      </div>
    );
  }
);

DateTimePicker.displayName = 'DateTimePicker';

export default DateTimePicker;
