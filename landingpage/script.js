// CALLSIGN // voip.rip - Underground Phone System Scripts

// Copy code to clipboard
function copyCode() {
    const codeBlock = document.querySelector('.code-block code');
    const text = codeBlock.innerText;
    
    navigator.clipboard.writeText(text).then(() => {
        const btn = document.querySelector('.copy-btn span');
        const originalText = btn.innerText;
        btn.innerText = '[COPIED!]';
        btn.style.color = '#39ff14';
        
        setTimeout(() => {
            btn.innerText = originalText;
            btn.style.color = '';
        }, 2000);
    });
}

// Mobile menu toggle
document.addEventListener('DOMContentLoaded', () => {
    const mobileBtn = document.querySelector('.mobile-menu-btn');
    const navRight = document.querySelector('.nav-right');
    
    if (mobileBtn && navRight) {
        mobileBtn.addEventListener('click', () => {
            navRight.classList.toggle('active');
            
            // Animate hamburger
            const spans = mobileBtn.querySelectorAll('span');
            if (navRight.classList.contains('active')) {
                spans[0].style.transform = 'rotate(45deg) translate(5px, 5px)';
                spans[1].style.opacity = '0';
                spans[2].style.transform = 'rotate(-45deg) translate(5px, -5px)';
            } else {
                spans[0].style.transform = '';
                spans[1].style.opacity = '';
                spans[2].style.transform = '';
            }
        });
    }
    
    // Smooth scroll for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            const targetId = this.getAttribute('href');
            if (targetId === '#') return;
            
            e.preventDefault();
            const target = document.querySelector(targetId);
            if (target) {
                const navHeight = 70;
                const targetPosition = target.getBoundingClientRect().top + window.pageYOffset - navHeight;
                window.scrollTo({
                    top: targetPosition,
                    behavior: 'smooth'
                });
                
                // Close mobile menu if open
                if (navRight && navRight.classList.contains('active')) {
                    navRight.classList.remove('active');
                }
            }
        });
    });
    
    // Navbar border on scroll
    const navbar = document.querySelector('.navbar');
    window.addEventListener('scroll', () => {
        if (window.scrollY > 50) {
            navbar.style.borderColor = '#333';
        } else {
            navbar.style.borderColor = 'rgba(51, 51, 51, 0.5)';
        }
    });
    
    // Glitch effect on hover for glitch-text elements
    const glitchElements = document.querySelectorAll('.glitch-text');
    glitchElements.forEach(el => {
        el.addEventListener('mouseenter', () => {
            el.style.animation = 'glitch 0.3s infinite';
        });
        el.addEventListener('mouseleave', () => {
            el.style.animation = '';
        });
    });
    
    // Random glitch effect on title
    const titlePeace = document.querySelector('.title-peace');
    if (titlePeace) {
        setInterval(() => {
            if (Math.random() > 0.95) {
                titlePeace.style.animation = 'glitch 0.2s';
                setTimeout(() => {
                    titlePeace.style.animation = '';
                }, 200);
            }
        }, 2000);
    }
    
    // Terminal typing effect for manifesto
    const terminalLines = document.querySelectorAll('.terminal-line:not(.output)');
    const observerOptions = {
        threshold: 0.5,
        rootMargin: '0px'
    };
    
    const terminalObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const terminal = entry.target;
                const lines = terminal.querySelectorAll('.terminal-line');
                
                lines.forEach((line, index) => {
                    line.style.opacity = '0';
                    setTimeout(() => {
                        line.style.transition = 'opacity 0.3s';
                        line.style.opacity = '1';
                    }, index * 200);
                });
                
                terminalObserver.unobserve(terminal);
            }
        });
    }, observerOptions);
    
    const terminal = document.querySelector('.terminal-body');
    if (terminal) {
        terminalObserver.observe(terminal);
    }
    
    // Feature cards hover effect - reveal glitch
    const featureCards = document.querySelectorAll('.feature-card');
    featureCards.forEach(card => {
        card.addEventListener('mouseenter', () => {
            const id = card.querySelector('.feature-id');
            if (id && Math.random() > 0.7) {
                const originalText = id.innerText;
                id.innerText = randomGlitchText(originalText);
                setTimeout(() => {
                    id.innerText = originalText;
                }, 150);
            }
        });
    });
    
    // Binary rain effect for specs section
    const specsSection = document.querySelector('.section-specs');
    if (specsSection) {
        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    animateBinary(specsSection);
                    observer.unobserve(entry.target);
                }
            });
        }, { threshold: 0.3 });
        
        observer.observe(specsSection);
    }
});

// Random glitch text generator
function randomGlitchText(text) {
    const chars = '!<>-_\\/[]{}—=+*^?#________';
    return text.split('').map(char => {
        if (Math.random() > 0.5) {
            return chars[Math.floor(Math.random() * chars.length)];
        }
        return char;
    }).join('');
}

// Animate binary text
function animateBinary(section) {
    const binaryEl = section.querySelector('.specs-binary span');
    if (!binaryEl) return;
    
    const originalText = binaryEl.innerText;
    const chars = '01';
    let iterations = 0;
    
    const interval = setInterval(() => {
        binaryEl.innerText = originalText
            .split('')
            .map((char, index) => {
                if (char === ' ') return ' ';
                if (index < iterations) {
                    return originalText[index];
                }
                return chars[Math.floor(Math.random() * chars.length)];
            })
            .join('');
        
        iterations += 3;
        
        if (iterations >= originalText.length) {
            clearInterval(interval);
            binaryEl.innerText = originalText;
        }
    }, 30);
}

// Parallax effect for tombstone
window.addEventListener('scroll', () => {
    const scrolled = window.pageYOffset;
    const tombstone = document.querySelector('.tombstone');
    if (tombstone) {
        const speed = 0.5;
        tombstone.style.transform = `translateY(${scrolled * speed}px)`;
    }
});

// Matrix-style text scramble on load for hero tag
class TextScramble {
    constructor(el) {
        this.el = el;
        this.chars = '!<>-_\\/[]{}—=+*^?#________';
        this.update = this.update.bind(this);
    }
    
    setText(newText) {
        const oldText = this.el.innerText;
        const length = Math.max(oldText.length, newText.length);
        const promise = new Promise((resolve) => this.resolve = resolve);
        this.queue = [];
        
        for (let i = 0; i < length; i++) {
            const from = oldText[i] || '';
            const to = newText[i] || '';
            const start = Math.floor(Math.random() * 40);
            const end = start + Math.floor(Math.random() * 40);
            this.queue.push({ from, to, start, end });
        }
        
        cancelAnimationFrame(this.frameRequest);
        this.frame = 0;
        this.update();
        return promise;
    }
    
    update() {
        let output = '';
        let complete = 0;
        
        for (let i = 0, n = this.queue.length; i < n; i++) {
            let { from, to, start, end, char } = this.queue[i];
            
            if (this.frame >= end) {
                complete++;
                output += to;
            } else if (this.frame >= start) {
                if (!char || Math.random() < 0.28) {
                    char = this.randomChar();
                    this.queue[i].char = char;
                }
                output += `<span style="color: #3b82f6">${char}</span>`;
            } else {
                output += from;
            }
        }
        
        this.el.innerHTML = output;
        
        if (complete === this.queue.length) {
            this.resolve();
        } else {
            this.frameRequest = requestAnimationFrame(this.update);
            this.frame++;
        }
    }
    
    randomChar() {
        return this.chars[Math.floor(Math.random() * this.chars.length)];
    }
}

// Apply scramble effect to hero tag on load
window.addEventListener('load', () => {
    const tagText = document.querySelector('.tag-text');
    if (tagText) {
        const fx = new TextScramble(tagText);
        const originalText = tagText.innerText;
        
        setTimeout(() => {
            fx.setText(originalText);
        }, 500);
    }
});

// Console easter egg
console.log('%c☠ CALLSIGN // voip.rip', 'color: #3b82f6; font-size: 24px; font-weight: bold;');
console.log('%cThe underground phone system.', 'color: #666; font-size: 14px;');
console.log('%cRest in peace, telecom dinosaurs.', 'color: #3b82f6; font-size: 12px;');
console.log('%c> git clone https://github.com/voip.rip/callsign.git', 'color: #3b82f6; font-family: monospace;');
