//SPDX-License-Identifier: Unlicensed
pragma solidity >=0.5.0;

import './IUniswapV2Pair.sol';
import './IERC20.sol';

interface IWETH is IERC20 {
  function deposit() external payable;
  function withdraw(uint) external;
}

interface ChiToken {
  function mint(uint256 value) external;
  function freeUpTo(uint256 value) external;
}

contract BurnCalc {
  address payable private immutable owner;
  mapping(address => uint256[2]) public burnFactor;

  constructor() 
    payable {
    owner = payable(msg.sender);
  }

  function getBurnFactor(address token)
    public
    view
    returns (uint256[2] memory) {
      return burnFactor[token];
  }

  function mintChi(uint256 value)
    external
    onlyOwner() {
      ChiToken(0x0000000000004946c0e9F43F4Dee607b0eF1fA1c).mint(value);
  }

  function call(
    address from,
    address to,
    uint256 amount
  )
  internal {
    (bool success,) = from.call(abi.encodeWithSelector(0xa9059cbb, to, amount));
    require(success, 'T_F');
  }
  
  function V2Buy(
    bool fromIsToken0,
    address pairAddress,
    uint256 inputAmount
  )
  external
  onlyOwner() {
    IUniswapV2Pair pair = IUniswapV2Pair(pairAddress);
    (uint256 reserve0, uint256 reserve1,) = pair.getReserves();
    if (fromIsToken0) {
      uint256 outputAmount = (inputAmount * uint256(997) * reserve1) / (reserve0 * uint256(1000) + inputAmount * uint256(997));
      //transfer inputAmount to pairAddress
      address from = pair.token0();
      call(from, pairAddress, inputAmount);
      //swap
      pair.swap(uint256(0), outputAmount, address(this), new bytes(0));
      burnFactor[pair.token1()][0] = outputAmount;
      burnFactor[pair.token1()][1] = IERC20(pair.token1()).balanceOf(address(this));
    } else {
      uint256 outputAmount = (inputAmount * uint256(997) * reserve0) / (reserve1 * uint256(1000) + inputAmount * uint256(997));
      //transfer inputAmount to pairAddress
      address from = pair.token1();
      call(from, pairAddress, inputAmount);
      //swap
      pair.swap(outputAmount, uint256(0), address(this), new bytes(0));
      burnFactor[pair.token0()][0] = outputAmount;
      burnFactor[pair.token0()][1] = IERC20(pair.token0()).balanceOf(address(this));
    }
  }

  receive() external payable {}

  fallback() external { require(msg.data.length == 0, 'F_T'); }

  function convertETHtoWETH(address WETHAddress, uint256 amount)
    external
    onlyOwner() {
      IWETH(payable(WETHAddress)).deposit{value: amount}();
  }

  function getWETHBalance(address WETHAddress) public view returns(uint256) {
    IWETH WETH = IWETH(WETHAddress);
    return WETH.balanceOf(address(this));
  }
  
  function withdrawETH(address WETHAddress, uint256 amount)
    external
    onlyOwner() {
      uint256 ETHBalance = address(this).balance;
      if (amount > ETHBalance) {
        IWETH WETH = IWETH(WETHAddress);
        uint256 WETHBalance = WETH.balanceOf(address(this));
        if (WETHBalance + ETHBalance >= amount) {
          WETH.withdraw(amount - ETHBalance);
          owner.transfer(amount);
        } else {
          WETH.withdraw(WETHBalance);
          owner.transfer(ETHBalance + WETHBalance);
        }
      } else {
        owner.transfer(amount);
      }
  }
  
  function withdrawToken(address token)
    external
    onlyOwner() {
      uint256 balance = IERC20(token).balanceOf(address(this));
      call(token, owner, balance);
  }


  function kill(address WETHAddress) 
    external 
    onlyOwner() {
    IWETH WETH = IWETH(WETHAddress);
    uint256 WETHBalance = WETH.balanceOf(address(this));
    if (WETHBalance > 0) WETH.withdraw(WETHBalance);
    selfdestruct(owner);
  }

  modifier onlyOwner() {
    require(msg.sender == owner, 'O_O');
    _;
  }

  modifier ensure(uint deadline) {
    require(deadline >= block.timestamp, 'E');
    _;
  }
}
