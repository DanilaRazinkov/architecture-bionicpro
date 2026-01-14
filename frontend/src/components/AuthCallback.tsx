import React, { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { createBFFAuthService } from '../auth/AuthService';

interface AuthCallbackProps {
  onSuccess?: (response: any) => void;
  onError?: (error: Error) => void;
}

export const AuthCallback: React.FC<AuthCallbackProps> = ({ onSuccess, onError }) => {
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [error, setError] = useState<string | null>(null);
  const processedRef = useRef(false);
  const navigate = useNavigate();

  useEffect(() => {
    if (processedRef.current) return;

    const handleCallback = async () => {
      processedRef.current = true;

      try {
        const params = new URLSearchParams(window.location.search);
        const code = params.get('code');
        const state = params.get('state');
        const errorParam = params.get('error');

        if (errorParam) {
          throw new Error(`OAuth error: ${params.get('error_description') || errorParam}`);
        }

        if (!code || !state) {
          throw new Error('Invalid authentication response');
        }

        const authService = createBFFAuthService();
        const response = await authService.handleCallback(code, state);

        if (!response.success) {
          throw new Error(response.message || 'Authentication failed');
        }

        setStatus('success');
        onSuccess?.(response);
        window.history.replaceState({}, document.title, window.location.pathname);

        setTimeout(() => navigate('/', { replace: true }), 1500);
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Unknown error';
        setError(message);
        setStatus('error');
        onError?.(err instanceof Error ? err : new Error(message));
      }
    };

    handleCallback();
  }, [onSuccess, onError, navigate]);

  if (status === 'loading') return <LoadingScreen />;
  if (status === 'success') return <SuccessScreen />;
  if (status === 'error') return <ErrorScreen error={error} onBack={() => navigate('/', { replace: true })} />;

  return null;
};

// Остальной код остается без изменений
const LoadingScreen = () => (
  <AuthScreen
    icon={<div className="auth-spinner"></div>}
    title="Completing login..."
    message="Processing authentication data"
  />
);

const SuccessScreen = () => (
  <AuthScreen
    icon={
      <div className="success-icon">
        <svg className="success-icon-svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
      </div>
    }
    title="Login successful"
    message="Redirecting to application..."
  />
);

interface ErrorScreenProps {
  error: string | null;
  onBack: () => void;
}

const ErrorScreen: React.FC<ErrorScreenProps> = ({ error, onBack }) => (
  <AuthScreen
    icon={
      <div className="error-icon">
        <svg className="error-icon-svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
      </div>
    }
    title="Login error"
    message={error || 'An unknown error occurred'}
    action={
      <button className="auth-button" onClick={onBack}>
        Back to home
      </button>
    }
  />
);

interface AuthScreenProps {
  icon: React.ReactNode;
  title: string;
  message: string;
  action?: React.ReactNode;
}

const AuthScreen: React.FC<AuthScreenProps> = ({ icon, title, message, action }) => (
  <div className="auth-container">
    <div className="auth-card">
      {icon}
      <h2 className="auth-title">{title}</h2>
      <p className="auth-message">{message}</p>
      {action && <div className="auth-action">{action}</div>}
    </div>
  </div>
);
