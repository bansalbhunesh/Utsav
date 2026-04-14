'use client'

import { useState } from 'react'

interface ConfettiPiece {
  id: number
  left: number
  delay: number
  scale: number
  color: string
}

export function Confetti() {
  const [pieces] = useState<ConfettiPiece[]>(() =>
    Array.from({ length: 50 }).map((_, i) => ({
      id: i,
      left: Math.random() * 100,
      delay: Math.random() * 2,
      scale: Math.random() * 0.5 + 0.5,
      color: ['#F97316', '#EA580C', '#FB923C', '#FDBA74', '#FFF'][Math.floor(Math.random() * 5)],
    }))
  )

  return (
    <div className="fixed inset-0 pointer-events-none z-[100] overflow-hidden">
      {pieces.map((p) => (
        <div
          key={p.id}
          className="absolute top-[-20px] w-2 h-4 rounded-full animate-confetti-fall"
          style={{
            left: `${p.left}%`,
            backgroundColor: p.color,
            animationDelay: `${p.delay}s`,
            transform: `scale(${p.scale})`,
          } as React.CSSProperties}
        />
      ))}
      <style jsx global>{`
        @keyframes confetti-fall {
          0% {
            transform: translateY(0) rotate(0deg);
            opacity: 1;
          }
          100% {
            transform: translateY(100vh) rotate(720deg);
            opacity: 0;
          }
        }
        .animate-confetti-fall {
          animation: confetti-fall 3s linear forwards;
        }
      `}</style>
    </div>
  )
}
