export type TIconType =
  | 'process'
  | 'network_device'
  | 'service'
  | 'FailurePoint'
  | 'physical_machine'
  | 'host'
  | 'pod'
  | 'middleware'
  | 'database'
  | '';

export interface TIcon {
  name: string;
  value: string;
  icon?: JSX.Element | string;
  type: TIconType;
}
