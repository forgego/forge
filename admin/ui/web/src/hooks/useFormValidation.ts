import { useMemo } from 'react';
import { z, ZodType } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import type { FieldMetadata } from '../api/types';

export interface UseFormValidationOptions {
  metadata: {
    fields: FieldMetadata[];
  };
  mode?: 'onChange' | 'onBlur' | 'onSubmit' | 'onTouched';
}

export function useFormValidation(options: UseFormValidationOptions) {
  const { metadata, mode = 'onChange' } = options;

  const schema = useMemo(() => {
    const shape: Record<string, ZodType<any>> = {};

    for (const field of metadata.fields) {
      if (field.read_only) continue;

      let validator: ZodType<any>;

      switch (field.type) {
        case 'string':
        case 'text':
        case 'email':
          validator = z.string();
          if (field.required) {
            validator = (validator as z.ZodString).min(1, `${field.label || field.name} is required`);
          } else {
            validator = (validator as z.ZodString).optional();
          }
          if (field.max_length) {
            validator = (validator as z.ZodString).max(field.max_length, `Maximum length is ${field.max_length}`);
          }
          if (field.min_length) {
            validator = (validator as z.ZodString).min(field.min_length, `Minimum length is ${field.min_length}`);
          }
          if (field.type === 'email') {
            validator = (validator as z.ZodString).email('Invalid email address');
          }
          break;

        case 'integer':
        case 'int64':
        case 'int32':
          validator = z.coerce.number();
          if (field.required) {
            validator = (validator as z.ZodNumber).min(0, `${field.label || field.name} is required`);
          } else {
            validator = (validator as z.ZodNumber).optional();
          }
          if (field.min_value !== undefined) {
            validator = (validator as z.ZodNumber).min(field.min_value);
          }
          if (field.max_value !== undefined) {
            validator = (validator as z.ZodNumber).max(field.max_value);
          }
          break;

        case 'float':
        case 'float64':
          validator = z.coerce.number();
          if (field.required) {
            validator = (validator as z.ZodNumber).min(0, `${field.label || field.name} is required`);
          } else {
            validator = (validator as z.ZodNumber).optional();
          }
          break;

        case 'boolean':
        case 'bool':
          validator = z.coerce.boolean();
          if (!field.required) {
            validator = (validator as z.ZodBoolean).optional();
          }
          break;

        case 'date':
          validator = z.string();
          if (field.required) {
            validator = (validator as z.ZodString).min(1, `${field.label || field.name} is required`);
          } else {
            validator = (validator as z.ZodString).optional();
          }
          break;

        case 'datetime':
          validator = z.string();
          if (field.required) {
            validator = (validator as z.ZodString).min(1, `${field.label || field.name} is required`);
          } else {
            validator = (validator as z.ZodString).optional();
          }
          break;

        case 'foreign_key':
        case 'relation':
          if (field.required) {
            validator = z.union([z.string(), z.number()]).refine((val) => val !== '', {
              message: `${field.label || field.name} is required`,
            });
          } else {
            validator = z.union([z.string(), z.number()]).optional();
          }
          break;

        case 'many_to_many':
        case 'array':
          validator = z.array(z.any());
          if (!field.required) {
            validator = (validator as z.ZodArray<any>).optional();
          }
          break;

        case 'json':
          validator = z.any();
          break;

        default:
          validator = z.string().optional();
      }

      shape[field.name] = validator;
    }

    return z.object(shape);
  }, [metadata.fields]);

  const form = useForm({
    resolver: zodResolver(schema),
    mode,
  });

  return form;
}

export function getFieldError(form: any, fieldName: string): string | undefined {
  const error = form.formState.errors[fieldName];
  if (error) {
    if (typeof error.message === 'string') {
      return error.message;
    }
    return `${fieldName} is invalid`;
  }
  return undefined;
}

export function hasFieldError(form: any, fieldName: string): boolean {
  return !!form.formState.errors[fieldName];
}
