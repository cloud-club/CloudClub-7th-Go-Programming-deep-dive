import React from 'react';
import '../styles/LoginExit.css';

function LoginExit({
  onLogin,
  onExit
}) {
  return (
    <div className="login-exit-main-container">
      <div className="login-exit-card">
        <div className="login-exit-left">
          <div className="login-exit-logo">💬 CloudClub</div>
          <div className="login-exit-title">Connect</div>
          <div className="login-exit-desc">
            Join chat rooms, share ideas,<br/>and make friends easily.
          </div>
          <div className="login-exit-button-group">
            <button className="login-exit-btn login" onClick={onLogin}>Login</button>
            <button className="login-exit-btn exit" onClick={onExit}>Exit</button>
          </div>
        </div>
        <div className="login-exit-right">
          {/* <img className="login-exit-illustration" src="(이미지 경로)" alt="Chat Illustration" /> */}
        </div>
      </div>
    </div>
  );
}

export default LoginExit;